package user

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/authctx"
	"github.com/vtech/our-sell/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c fiber.Ctx) error {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	ctx, cancel := requestContext(c)
	defer cancel()
	result, err := h.service.List(ctx, c.Query("search"), c.Query("role"), page, pageSize)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: result})
}

func (h *Handler) UpdateRole(c fiber.Ctx) error {
	identity, ok := c.Locals(authctx.LocalsKey).(authctx.Context)
	if !ok || identity.UserID == "" {
		return apperror.ErrUnauthorized
	}
	var req UpdateRoleRequest
	if err := response.DecodeJSON(c.Body(), &req); err != nil {
		return apperror.NewValidation(map[string]string{"body": "must be valid JSON with no unknown fields"})
	}
	ctx, cancel := requestContext(c)
	defer cancel()
	updated, err := h.service.UpdateRole(ctx, identity.UserID, c.Params("userID"), req.Role)
	if err != nil {
		return err
	}
	return c.JSON(response.Envelope{Data: updated.Public()})
}

func queryInt(c fiber.Ctx, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func requestContext(c fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 10*time.Second)
}
