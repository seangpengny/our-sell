package middleware

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func Recovery(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(c.Context(), "panic recovered", "request_id", RequestIDFrom(c))
				err = fmt.Errorf("panic recovered: %v", recovered)
			}
		}()
		return c.Next()
	}
}
