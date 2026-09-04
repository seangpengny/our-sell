package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
	"github.com/vtech/our-sell/internal/platform/response"
)

type CookieConfig struct {
	Name     string
	Domain   string
	Secure   bool
	SameSite string
	MaxAge   int
}

type Handler struct {
	service *Service
	cookie  CookieConfig
}

func NewHandler(service *Service, cookie CookieConfig) *Handler {
	return &Handler{service: service, cookie: cookie}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	account, err := h.service.Register(ctx, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(response.Envelope{Data: account.Public()})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	result, err := h.service.Login(ctx, req, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, result.RefreshToken)
	return c.JSON(response.Envelope{Data: LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
		User:        result.User.Public(),
	}})
}

func (h *Handler) Refresh(c fiber.Ctx) error {
	rawToken, found := readCookie(c, h.cookie.Name)
	if !found {
		var req RefreshRequest
		if err := decodeBody(c, &req); err != nil {
			return err
		}
		rawToken = req.RefreshToken
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	result, err := h.service.Refresh(ctx, rawToken)
	if err != nil {
		h.clearRefreshCookie(c)
		return err
	}
	h.setRefreshCookie(c, result.RefreshToken)
	return c.JSON(response.Envelope{Data: LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
		User:        result.User.Public(),
	}})
}

func (h *Handler) Logout(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.Logout(ctx, identity.SessionID); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return c.JSON(response.Message{Message: "logged out"})
}

func (h *Handler) LogoutAll(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.LogoutAll(ctx, identity.UserID); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return c.JSON(response.Message{Message: "all sessions revoked"})
}

func (h *Handler) ForgotPassword(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.ForgotPassword(ctx, req.Email); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "If an account exists for that email, password reset instructions have been sent."})
}

func (h *Handler) ResetPassword(c fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.ResetPassword(ctx, req); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return c.JSON(response.Message{Message: "password reset successfully"})
}

func (h *Handler) VerifyEmail(c fiber.Ctx) error {
	var req TokenRequest
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.VerifyEmail(ctx, req.Token); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "email verified"})
}

func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.ResendVerification(ctx, req.Email); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "If an account exists and is not verified, a verification email has been sent."})
}

func (h *Handler) ChangePassword(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	var req ChangePasswordRequest
	if err := decodeBody(c, &req); err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.ChangePassword(ctx, identity.UserID, identity.SessionID, req); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "password changed successfully"})
}

func (h *Handler) Me(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	account, err := h.service.Me(ctx, identity.UserID)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: account.Public()})
}

func (h *Handler) Sessions(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	sessions, err := h.service.Sessions(ctx, identity.UserID)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: sessions})
}

func (h *Handler) RevokeSession(c fiber.Ctx) error {
	identity, err := identityFrom(c)
	if err != nil {
		return err
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	if err := h.service.RevokeUserSession(ctx, identity.UserID, c.Params("sessionID")); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "session revoked"})
}

func decodeBody(c fiber.Ctx, destination any) error {
	if err := response.DecodeJSON(c.Body(), destination); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	return nil
}

func identityFrom(c fiber.Ctx) (authctx.Context, error) {
	value := c.Locals(authctx.LocalsKey)
	identity, ok := value.(authctx.Context)
	if !ok || identity.UserID == "" || identity.SessionID == "" {
		return authctx.Context{}, apperror.ErrUnauthorized
	}
	return identity, nil
}

func requestContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 10*time.Second)
}

func (h *Handler) setRefreshCookie(c fiber.Ctx, raw string) {
	cookie := http.Cookie{Name: h.cookie.Name, Value: raw, Path: "/api/v1/auth", Domain: h.cookie.Domain, MaxAge: h.cookie.MaxAge, HttpOnly: true, Secure: h.cookie.Secure, SameSite: sameSite(h.cookie.SameSite)}
	c.Set("Set-Cookie", cookie.String())
}

func (h *Handler) clearRefreshCookie(c fiber.Ctx) {
	cookie := http.Cookie{Name: h.cookie.Name, Value: "", Path: "/api/v1/auth", Domain: h.cookie.Domain, MaxAge: -1, HttpOnly: true, Secure: h.cookie.Secure, SameSite: sameSite(h.cookie.SameSite), Expires: time.Unix(1, 0)}
	c.Set("Set-Cookie", cookie.String())
}

func readCookie(c fiber.Ctx, name string) (string, bool) {
	for _, part := range strings.Split(c.Get("Cookie"), ";") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if found && key == name && value != "" {
			return value, true
		}
	}
	return "", false
}

func sameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
