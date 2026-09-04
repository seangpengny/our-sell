package middleware

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
)

func Logger(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			status = errorStatus(err)
		}
		attrs := []any{
			"request_id", RequestIDFrom(c),
			"method", c.Method(),
			"route", c.Path(),
			"status", status,
			"duration_ms", time.Since(started).Milliseconds(),
			"client_ip", c.IP(),
		}
		if identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context); ok {
			attrs = append(attrs, "user_id", identity.UserID)
		}
		if err != nil {
			attrs = append(attrs, "error", err)
			logger.ErrorContext(c.Context(), "http request failed", attrs...)
			return err
		}
		logger.InfoContext(c.Context(), "http request", attrs...)
		return nil
	}
}

func errorStatus(err error) int {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code
	}
	return apperror.Public(err).Status
}
