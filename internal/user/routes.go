package user

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *Handler, requireAuth, requireAdmin fiber.Handler) {
	admin := router.Group("/admin/users", requireAuth, requireAdmin)
	admin.Get("", handler.List)
	admin.Patch("/:userID/role", handler.UpdateRole)
}
