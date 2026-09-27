package facebook

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
	"github.com/vtech/our-sell/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func RegisterRoutes(router fiber.Router, handler *Handler, requireAuth, requireAdmin fiber.Handler) {
	auth := router.Group("/auth/facebook")
	auth.Post("/connect", requireAuth, handler.Connect)
	auth.Get("/callback", handler.Callback)

	facebook := router.Group("/facebook", requireAuth)
	facebook.Get("/connections", handler.Connections)
	facebook.Get("/pages", handler.Pages)
	facebook.Get("/page-inventory", requireAdmin, handler.Inventory)
	facebook.Put("/pages/:pageID/listing", requireAdmin, handler.SaveListing)
	facebook.Post("/connections/:connectionID/sync", requireAdmin, handler.Sync)
	facebook.Delete("/connections/:connectionID", handler.Disconnect)
	router.Post("/marketplace/pages/:listingID/orders", requireAuth, handler.CreateOrder)
	router.Get("/marketplace/pages", handler.PublicListings)
	router.Get("/marketplace/pages/:listingID", handler.PublicListing)
}

func (h *Handler) Connect(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" || identity.SessionID == "" {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	authorizationURL, err := h.service.Connect(ctx, identity.UserID, identity.SessionID)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: struct {
		AuthorizationURL string `json:"authorization_url"`
	}{AuthorizationURL: authorizationURL}})
}

func (h *Handler) Callback(c fiber.Ctx) error {
	ctx, cancel := handlerContext(c)
	defer cancel()
	status, err := h.service.Callback(ctx, CallbackParams{
		Code:        trimOAuthValue(c.Query("code")),
		State:       trimOAuthValue(c.Query("state")),
		OAuthError:  trimOAuthValue(c.Query("error")),
		ErrorReason: trimOAuthValue(c.Query("error_reason")),
	})
	if err != nil {
		status = CallbackError
	}
	return c.Redirect().To(h.service.redirectURL(status))
}

func (h *Handler) Connections(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	connections, err := h.service.Connections(ctx, identity.UserID)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: connections})
}

func (h *Handler) Pages(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	pages, err := h.service.Pages(ctx, identity.UserID, c.Query("connection_id"))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: pages})
}

func (h *Handler) Inventory(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	connectionID := uuidOrNil(c.Query("connection_id"))
	ctx, cancel := handlerContext(c)
	defer cancel()
	result, err := h.service.StoredPages(ctx, identity.UserID, pageInventoryParams{
		ConnectionID: connectionID, Search: trimOAuthValue(c.Query("search")),
		Page: queryInt(c, "page", 1), PageSize: queryInt(c, "page_size", 25),
	})
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: result})
}

func (h *Handler) SaveListing(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	var input ListingInput
	if err := response.DecodeJSON(c.Body(), &input); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	listing, err := h.service.SaveListing(ctx, identity.UserID, c.Params("pageID"), input)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: listing})
}

func (h *Handler) Sync(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	pages, err := h.service.SyncConnection(ctx, identity.UserID, c.Params("connectionID"))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: struct {
		PageCount int `json:"page_count"`
	}{PageCount: len(pages)}})
}

func (h *Handler) PublicListings(c fiber.Ctx) error {
	ctx, cancel := handlerContext(c)
	defer cancel()
	result, err := h.service.PublishedListings(ctx, c.Query("search"), queryInt(c, "page", 1), queryInt(c, "page_size", 25))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: result})
}

func (h *Handler) PublicListing(c fiber.Ctx) error {
	ctx, cancel := handlerContext(c)
	defer cancel()
	listing, err := h.service.PublishedListing(ctx, c.Params("listingID"))
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: listing})
}

func (h *Handler) CreateOrder(c fiber.Ctx) error {
	var input MarketplaceOrderInput
	if err := response.DecodeJSON(c.Body(), &input); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	order, err := h.service.CreateOrder(ctx, identity.UserID, c.Params("listingID"), input)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(response.Envelope{Data: order})
}

func (h *Handler) Disconnect(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	ctx, cancel := handlerContext(c)
	defer cancel()
	if err := h.service.Disconnect(ctx, identity.UserID, c.Params("connectionID")); err != nil {
		return err
	}
	return c.JSON(response.Message{Message: "Facebook account disconnected"})
}

func handlerContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 30*time.Second)
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func uuidOrNil(value string) uuid.UUID {
	if value == "" {
		return uuid.Nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}
