package facebook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientConfig struct {
	AppID       string
	AppSecret   string
	RedirectURI string
	Version     string
	HTTPClient  *http.Client
	BaseURL     string
}

type Client struct {
	appID       string
	appSecret   string
	redirectURI string
	version     string
	httpClient  *http.Client
	baseURL     *url.URL
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type profileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type pagesResponse struct {
	Data   []Page `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type graphErrorResponse struct {
	Error struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"code"`
		ErrorSubcode int    `json:"error_subcode"`
		IsTransient  bool   `json:"is_transient"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

type GraphError struct {
	Status    int
	Code      int
	Subcode   int
	Message   string
	Transient bool
}

func (e *GraphError) Error() string {
	return fmt.Sprintf("meta Graph API request failed (status=%d code=%d subcode=%d)", e.Status, e.Code, e.Subcode)
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.AppID == "" || cfg.AppSecret == "" || cfg.RedirectURI == "" || cfg.Version == "" {
		return nil, errors.New("facebook client configuration is incomplete")
	}
	base := cfg.BaseURL
	if base == "" {
		base = "https://graph.facebook.com"
	}
	parsedBase, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return nil, errors.New("facebook Graph API base URL is invalid")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		appID: cfg.AppID, appSecret: cfg.AppSecret, redirectURI: cfg.RedirectURI,
		version: strings.Trim(cfg.Version, "/"), httpClient: cfg.HTTPClient, baseURL: parsedBase,
	}, nil
}

func (c *Client) AuthorizationURL(state string) (string, error) {
	if state == "" {
		return "", errors.New("facebook OAuth state is required")
	}
	endpoint := c.baseURLFor("https://www.facebook.com")
	endpoint.Path = "/" + c.version + "/dialog/oauth"
	query := endpoint.Query()
	query.Set("client_id", c.appID)
	query.Set("redirect_uri", c.redirectURI)
	query.Set("state", state)
	query.Set("scope", requestedScope)
	query.Set("response_type", "code")
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (c *Client) ExchangeAuthorizationCode(ctx context.Context, code string) (accessTokenResponse, error) {
	if code == "" || len(code) > 4096 {
		return accessTokenResponse{}, errors.New("authorization code is invalid")
	}
	query := url.Values{}
	query.Set("client_id", c.appID)
	query.Set("client_secret", c.appSecret)
	query.Set("redirect_uri", c.redirectURI)
	query.Set("code", code)
	shortLived, err := c.getToken(ctx, "/oauth/access_token", query)
	if err != nil {
		return accessTokenResponse{}, err
	}
	if shortLived.AccessToken == "" {
		return accessTokenResponse{}, errors.New("meta returned an empty access token")
	}

	// Facebook Login supports exchanging the short-lived user token for a
	// longer-lived user token. Keep the app secret on this server-only path.
	longLivedQuery := url.Values{}
	longLivedQuery.Set("grant_type", "fb_exchange_token")
	longLivedQuery.Set("client_id", c.appID)
	longLivedQuery.Set("client_secret", c.appSecret)
	longLivedQuery.Set("fb_exchange_token", shortLived.AccessToken)
	longLived, err := c.getToken(ctx, "/oauth/access_token", longLivedQuery)
	if err != nil {
		return accessTokenResponse{}, err
	}
	if longLived.AccessToken == "" {
		return accessTokenResponse{}, errors.New("meta returned an empty long-lived access token")
	}
	return longLived, nil
}

func (c *Client) GetProfile(ctx context.Context, accessToken string) (profileResponse, error) {
	var profile profileResponse
	if err := c.getJSON(ctx, "/me", url.Values{"fields": []string{"id,name"}}, accessToken, &profile); err != nil {
		return profileResponse{}, err
	}
	if profile.ID == "" || profile.Name == "" || len(profile.ID) > 128 || len(profile.Name) > 255 {
		return profileResponse{}, errors.New("meta returned an invalid Facebook profile")
	}
	return profile, nil
}

func (c *Client) ListPages(ctx context.Context, accessToken string) ([]Page, error) {
	next := "/me/accounts"
	pages := make([]Page, 0)
	seen := make(map[string]struct{})
	for requestCount := 0; next != "" && requestCount < maxPageRequests; requestCount++ {
		if _, exists := seen[next]; exists {
			return nil, errors.New("meta pagination repeated the same page")
		}
		seen[next] = struct{}{}

		var response pagesResponse
		query := url.Values{"fields": []string{"id,name"}, "limit": []string{"100"}}
		if strings.HasPrefix(next, "http") {
			parsed, err := c.validateNextURL(next)
			if err != nil {
				return nil, err
			}
			query = parsed.Query()
			query.Del("access_token")
			next = parsed.Path
			if parsed.RawQuery != "" {
				next += "?" + query.Encode()
			}
		}
		if err := c.getJSON(ctx, next, query, accessToken, &response); err != nil {
			return nil, err
		}
		for _, page := range response.Data {
			if page.ID == "" || page.Name == "" || len(page.ID) > 128 || len(page.Name) > 255 {
				return nil, errors.New("meta returned an invalid Facebook Page")
			}
			pages = append(pages, page)
		}
		next = response.Paging.Next
	}
	if next != "" {
		return nil, errors.New("meta returned too many Page result pages")
	}
	return pages, nil
}

func (c *Client) getToken(ctx context.Context, path string, query url.Values) (accessTokenResponse, error) {
	var response accessTokenResponse
	if err := c.getJSON(ctx, path, query, "", &response); err != nil {
		return accessTokenResponse{}, err
	}
	return response, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, accessToken string, destination any) error {
	endpoint, err := c.endpoint(path)
	if err != nil {
		return err
	}
	if query == nil {
		query = url.Values{}
	}
	if accessToken != "" {
		query.Set("access_token", accessToken)
	}
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create Meta request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("request Meta Graph API: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body := io.LimitReader(response.Body, maxMetaResponseBytes)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var graphResponse graphErrorResponse
		_ = json.NewDecoder(body).Decode(&graphResponse)
		return &GraphError{
			Status: response.StatusCode, Code: graphResponse.Error.Code,
			Subcode: graphResponse.Error.ErrorSubcode, Message: graphResponse.Error.Message,
			Transient: graphResponse.Error.IsTransient,
		}
	}
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode Meta response: %w", err)
	}
	return nil
}

func (c *Client) endpoint(path string) (*url.URL, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return c.validateNextURL(path)
	}
	if path == "" || !strings.HasPrefix(path, "/") || strings.Contains(path, "//") {
		return nil, errors.New("meta API path is invalid")
	}
	endpoint := *c.baseURL
	if strings.HasPrefix(path, "/"+c.version+"/") {
		endpoint.Path = path
	} else {
		endpoint.Path = "/" + c.version + path
	}
	return &endpoint, nil
}

func (c *Client) baseURLFor(raw string) *url.URL {
	parsed, _ := url.Parse(raw)
	return parsed
}

func (c *Client) validateNextURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != c.baseURL.Scheme || parsed.Host != c.baseURL.Host || parsed.User != nil {
		return nil, errors.New("meta pagination URL is invalid")
	}
	versionPrefix := "/" + c.version + "/"
	if !strings.HasPrefix(parsed.Path, versionPrefix) {
		return nil, errors.New("meta pagination URL has an unexpected path")
	}
	return parsed, nil
}

func IsInvalidToken(err error) bool {
	var graphErr *GraphError
	return errors.As(err, &graphErr) && (graphErr.Code == 190 || graphErr.Status == http.StatusUnauthorized)
}

func IsPermissionError(err error) bool {
	var graphErr *GraphError
	return errors.As(err, &graphErr) && (graphErr.Code == 10 || graphErr.Code == 200 || graphErr.Code == 2500)
}
