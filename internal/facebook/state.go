package facebook

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StateStore interface {
	Save(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Consume(ctx context.Context, key string) ([]byte, error)
}

type redisStateStore struct {
	client *redis.Client
}

var consumeStateScript = redis.NewScript(`
local value = redis.call('GET', KEYS[1])
if value then
  redis.call('DEL', KEYS[1])
end
return value
`)

func NewRedisStateStore(client *redis.Client) StateStore {
	return &redisStateStore{client: client}
}

func (s *redisStateStore) Save(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *redisStateStore) Consume(ctx context.Context, key string) ([]byte, error) {
	value, err := consumeStateScript.Run(ctx, s.client, []string{key}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, errStateNotFound
	}
	if err != nil {
		return nil, err
	}
	text, ok := value.(string)
	if !ok {
		return nil, errors.New("oauth state has an invalid Redis value")
	}
	return []byte(text), nil
}

var errStateNotFound = errors.New("oauth state not found")

func newOAuthState() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return hex.EncodeToString(value), nil
}

func stateKey(state string) string {
	digest := sha256.Sum256([]byte(state))
	return "facebook:oauth:state:" + hex.EncodeToString(digest[:])
}

func encodeOAuthState(value oauthState) ([]byte, error) {
	return json.Marshal(value)
}

func decodeOAuthState(value []byte) (oauthState, error) {
	var state oauthState
	if err := json.Unmarshal(value, &state); err != nil || state.UserID == "" || state.SessionID == "" {
		return oauthState{}, errors.New("invalid oauth state payload")
	}
	return state, nil
}
