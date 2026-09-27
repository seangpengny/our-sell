package facebook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListPagesFollowsValidatedPagination(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("access_token") != "user-token" {
			t.Fatalf("access token was not sent server-side")
		}
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("after") == "second" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"data": []Page{{ID: "2", Name: "Second Page"}}})
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data":   []Page{{ID: "1", Name: "First Page"}},
			"paging": map[string]string{"next": server.URL + "/v26.0/me/accounts?after=second&access_token=should-not-be-reused"},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		AppID: "app-id", AppSecret: "app-secret", RedirectURI: "http://localhost/callback",
		Version: "v26.0", BaseURL: server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	pages, err := client.ListPages(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("ListPages() error = %v", err)
	}
	if len(pages) != 2 || pages[0].ID != "1" || pages[1].ID != "2" {
		t.Fatalf("unexpected pages: %#v", pages)
	}
}

func TestListPagesRejectsForeignPaginationURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data":   []Page{{ID: "1", Name: "First Page"}},
			"paging": map[string]string{"next": "https://evil.example/v26.0/me/accounts?access_token=leak"},
		})
	}))
	defer server.Close()
	client, err := NewClient(ClientConfig{AppID: "id", AppSecret: "secret", RedirectURI: "http://localhost/callback", Version: "v26.0", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListPages(context.Background(), "token"); err == nil || !strings.Contains(err.Error(), "pagination URL") {
		t.Fatalf("expected pagination URL validation error, got %v", err)
	}
}

func TestExchangeAuthorizationCodeUsesLongLivedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		writer.Header().Set("Content-Type", "application/json")
		if query.Get("grant_type") == "fb_exchange_token" {
			if query.Get("fb_exchange_token") != "short-token" {
				t.Fatalf("short-lived token was not exchanged")
			}
			_ = json.NewEncoder(writer).Encode(accessTokenResponse{AccessToken: "long-token", ExpiresIn: 5184000})
			return
		}
		if query.Get("code") != "authorization-code" || query.Get("client_secret") != "app-secret" {
			t.Fatalf("authorization code exchange query was incomplete")
		}
		_ = json.NewEncoder(writer).Encode(accessTokenResponse{AccessToken: "short-token", ExpiresIn: 3600})
	}))
	defer server.Close()
	client, err := NewClient(ClientConfig{AppID: "app-id", AppSecret: "app-secret", RedirectURI: "http://localhost/callback", Version: "v26.0", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ExchangeAuthorizationCode(context.Background(), "authorization-code")
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken != "long-token" || result.ExpiresIn != 5184000 {
		t.Fatalf("unexpected exchanged token: %#v", result)
	}
}
