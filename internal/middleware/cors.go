package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func CORS(allowedOrigins []string) fiber.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; !ok {
				return fiber.NewError(fiber.StatusForbidden, "origin is not allowed")
			}
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Vary", "Origin")
			c.Set("Access-Control-Allow-Credentials", "true")
		}
		c.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if strings.EqualFold(c.Method(), fiber.MethodOptions) {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}
