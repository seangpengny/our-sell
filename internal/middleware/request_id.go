package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gofiber/fiber/v3"
)

const requestIDKey = "request_id"

var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if !safeRequestID.MatchString(id) {
			id = newRequestID()
		}
		c.Locals(requestIDKey, id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

func RequestIDFrom(c fiber.Ctx) string {
	id, _ := c.Locals(requestIDKey).(string)
	return id
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(value)
}
