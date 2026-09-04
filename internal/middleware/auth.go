package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/vtech/our-sell/internal/auth"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
)

func RequireAuth(tokens *auth.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := strings.TrimSpace(c.Get("Authorization"))
		scheme, raw, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || raw == "" {
			return apperror.ErrUnauthorized
		}
		claims, err := tokens.ValidateAccessToken(strings.TrimSpace(raw))
		if err != nil {
			return apperror.ErrUnauthorized
		}
		c.Locals(authctx.LocalsKey, authctx.Context{UserID: claims.Subject, SessionID: claims.SessionID, Role: claims.Role})
		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
		if !ok || identity.Role != role {
			return apperror.ErrForbidden
		}
		return c.Next()
	}
}
