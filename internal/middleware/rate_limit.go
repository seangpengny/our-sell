package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"github.com/vtech/our-sell/internal/platform/apperror"
)

type rateLimit struct {
	max    int
	window time.Duration
}

var rateLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func RateLimit(client *redis.Client, failOpen bool) fiber.Handler {
	limits := map[string]rateLimit{
		"/api/v1/auth/login":               {max: 10, window: 15 * time.Minute},
		"/api/v1/auth/register":            {max: 5, window: 15 * time.Minute},
		"/api/v1/auth/forgot-password":     {max: 5, window: 15 * time.Minute},
		"/api/v1/auth/reset-password":      {max: 10, window: 15 * time.Minute},
		"/api/v1/auth/resend-verification": {max: 5, window: 15 * time.Minute},
	}
	return func(c fiber.Ctx) error {
		if client == nil {
			if failOpen {
				return c.Next()
			}
			return apperror.ErrDependencyUnavailable
		}
		limit, ok := limits[c.Path()]
		if !ok {
			limit = rateLimit{max: 120, window: time.Minute}
		}
		identity := strings.TrimSpace(c.IP())
		if identity == "" {
			identity = "unknown"
		}
		key := fmt.Sprintf("rate:%s:%x", c.Path(), sha256.Sum256([]byte(identity)))
		if emailHash := requestEmailHash(c.Body()); emailHash != "" {
			key += ":email:" + emailHash
		}
		count, err := rateLimitScript.Run(c.Context(), client, []string{key}, limit.window.Milliseconds()).Int()
		if err != nil {
			if failOpen {
				return c.Next()
			}
			return apperror.ErrDependencyUnavailable
		}
		if count > limit.max {
			c.Set("Retry-After", strconv.Itoa(int(limit.window.Seconds())))
			return apperror.ErrRateLimited
		}
		return c.Next()
	}
}

func requestEmailHash(body []byte) string {
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || strings.TrimSpace(payload.Email) == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(payload.Email))))
	return hex.EncodeToString(digest[:])
}
