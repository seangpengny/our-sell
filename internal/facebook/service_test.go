package facebook

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeStateStore struct{ values map[string][]byte }

func (s *fakeStateStore) Save(_ context.Context, key string, value []byte, _ time.Duration) error {
	if s.values == nil {
		s.values = make(map[string][]byte)
	}
	s.values[key] = append([]byte(nil), value...)
	return nil
}

func (s *fakeStateStore) Consume(_ context.Context, key string) ([]byte, error) {
	value, ok := s.values[key]
	if !ok {
		return nil, errStateNotFound
	}
	delete(s.values, key)
	return value, nil
}

type fakeRepository struct{ saved storedConnection }

func (r *fakeRepository) Upsert(_ context.Context, userID uuid.UUID, facebookUserID, name string, encryptedToken []byte, expiresAt *time.Time, scopes string) (storedConnection, error) {
	r.saved = storedConnection{Connection: Connection{ID: "connection-id", FacebookUserID: facebookUserID, Name: name, TokenExpiresAt: expiresAt}, UserID: userID.String(), EncryptedAccessToken: encryptedToken, Scopes: scopes}
	return r.saved, nil
}

func (r *fakeRepository) ListActive(context.Context, uuid.UUID) ([]storedConnection, error) {
	return nil, nil
}
func (r *fakeRepository) GetActive(context.Context, uuid.UUID, uuid.UUID) (storedConnection, error) {
	return r.saved, nil
}
func (r *fakeRepository) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func TestCallbackStoresConnectionAndRejectsStateReplay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/v26.0/oauth/access_token":
			if request.URL.Query().Get("grant_type") == "fb_exchange_token" {
				_ = json.NewEncoder(writer).Encode(accessTokenResponse{AccessToken: "long-token", ExpiresIn: 3600})
			} else {
				_ = json.NewEncoder(writer).Encode(accessTokenResponse{AccessToken: "short-token", ExpiresIn: 3600})
			}
		case "/v26.0/me":
			_ = json.NewEncoder(writer).Encode(profileResponse{ID: "fb-user-1", Name: "Example User"})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	client, err := NewClient(ClientConfig{AppID: "id", AppSecret: "secret", RedirectURI: "http://localhost/callback", Version: "v26.0", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStateStore{}
	repository := &fakeRepository{}
	cipher, err := NewTokenCipher([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(repository, client, cipher, store, "http://localhost:3000", slog.Default())
	userID := uuid.New().String()
	authorizationURL, err := service.Connect(context.Background(), userID, uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	state := parsed.Query().Get("state")
	status, err := service.Callback(context.Background(), CallbackParams{State: state, Code: "code"})
	if err != nil || status != CallbackConnected {
		t.Fatalf("Callback() = %q, %v", status, err)
	}
	if repository.saved.UserID != userID || repository.saved.FacebookUserID != "fb-user-1" || len(repository.saved.EncryptedAccessToken) == 0 {
		t.Fatalf("connection was not safely stored: %#v", repository.saved)
	}
	status, err = service.Callback(context.Background(), CallbackParams{State: state, Code: "code"})
	if err != nil || status != CallbackStateErr {
		t.Fatalf("replayed Callback() = %q, %v", status, err)
	}
}

func TestCallbackMapsFacebookCancellation(t *testing.T) {
	store := &fakeStateStore{}
	repository := &fakeRepository{}
	cipher, _ := NewTokenCipher([]byte("01234567890123456789012345678901"))
	client, _ := NewClient(ClientConfig{AppID: "id", AppSecret: "secret", RedirectURI: "http://localhost/callback", Version: "v26.0"})
	service := NewService(repository, client, cipher, store, "http://localhost:3000", slog.Default())
	authorizationURL, err := service.Connect(context.Background(), uuid.New().String(), uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(authorizationURL)
	status, err := service.Callback(context.Background(), CallbackParams{State: parsed.Query().Get("state"), OAuthError: "access_denied"})
	if err != nil || status != CallbackCancelled {
		t.Fatalf("cancelled Callback() = %q, %v", status, err)
	}
	if strings.Contains(service.redirectURL(status), "access_token") || errors.Is(err, errStateNotFound) {
		t.Fatal("callback leaked token state")
	}
}
