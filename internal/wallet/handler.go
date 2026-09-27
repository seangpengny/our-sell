package wallet

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
	"github.com/vtech/our-sell/internal/platform/response"
)

type Handler struct {
	service       *Service
	webhookSecret string
}

func NewHandler(service *Service, webhookSecret string) *Handler {
	return &Handler{service: service, webhookSecret: webhookSecret}
}

func RegisterRoutes(router fiber.Router, handler *Handler, requireAuth, requireAdmin fiber.Handler) {
	wallet := router.Group("/wallet", requireAuth)
	wallet.Get("", handler.Wallet)
	wallet.Get("/ledger", handler.Ledger)
	wallet.Post("/topups", handler.CreateTopup)
	wallet.Get("/topups/:topupID", handler.GetTopup)
	router.Post("/payments/bakong/webhook", handler.Webhook)

	admin := router.Group("/admin/wallet", requireAuth, requireAdmin)
	admin.Get("/topups", handler.AdminTopups)
	admin.Post("/topups/:topupID/recheck", handler.AdminRecheckTopup)
}

func (h *Handler) Wallet(c fiber.Ctx) error {
	identity, ok := identity(c)
	if !ok {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	wallet, err := h.service.Wallet(ctx, identity.UserID)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: wallet})
}

func (h *Handler) Ledger(c fiber.Ctx) error {
	identity, ok := identity(c)
	if !ok {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	ledger, err := h.service.Ledger(ctx, identity.UserID, queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: ledger})
}

func (h *Handler) CreateTopup(c fiber.Ctx) error {
	identity, ok := identity(c)
	if !ok {
		return apperror.ErrUnauthorized
	}
	var input TopupInput
	if err := response.DecodeJSON(c.Body(), &input); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	topup, err := h.service.CreateTopup(ctx, identity.UserID, input)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(response.Envelope{Data: topup})
}

func (h *Handler) GetTopup(c fiber.Ctx) error {
	identity, ok := identity(c)
	if !ok {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	topup, err := h.service.GetTopup(ctx, identity.UserID, c.Params("topupID"))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: topup})
}

func (h *Handler) Webhook(c fiber.Ctx) error {
	if h.webhookSecret == "" {
		return apperror.ErrDependencyUnavailable
	}
	provided := c.Get("X-Bakong-Webhook-Secret")
	if len(provided) != len(h.webhookSecret) || subtle.ConstantTimeCompare([]byte(provided), []byte(h.webhookSecret)) != 1 {
		return apperror.ErrUnauthorized
	}
	var input WebhookInput
	if err := response.DecodeJSON(c.Body(), &input); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	if _, err := h.service.HandleWebhook(ctx, input); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "Bakong payment verification accepted"})
}

func (h *Handler) AdminTopups(c fiber.Ctx) error {
	ctx, cancel := handlerContext(c)
	defer cancel()
	result, err := h.service.AdminTopups(ctx, c.Query("search"), c.Query("status"), queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: result})
}

func (h *Handler) AdminRecheckTopup(c fiber.Ctx) error {
	ctx, cancel := handlerContext(c)
	defer cancel()
	topup, err := h.service.AdminRecheckTopup(ctx, c.Params("topupID"))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: topup})
}

func identity(c fiber.Ctx) (authctx.Context, bool) {
	value, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	return value, ok && value.UserID != ""
}

func handlerContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 30*time.Second)
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return fallback
	}
	parsed := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return fallback
		}
		parsed = parsed*10 + int(char-'0')
		if parsed > 1000 {
			return fallback
		}
	}
	if parsed == 0 {
		return fallback
	}
	return parsed
}
