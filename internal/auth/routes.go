package auth

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler, requireAuth fiber.Handler) {
	group := router.Group("/auth")
	group.Post("/register", handler.Register)
	group.Post("/login", handler.Login)
	group.Post("/refresh", handler.Refresh)
	group.Post("/forgot-password", handler.ForgotPassword)
	group.Post("/reset-password", handler.ResetPassword)
	group.Post("/verify-email", handler.VerifyEmail)
	group.Post("/resend-verification", handler.ResendVerification)

	group.Get("/me", requireAuth, handler.Me)
	group.Post("/logout", requireAuth, handler.Logout)
	group.Post("/logout-all", requireAuth, handler.LogoutAll)
	group.Post("/change-password", requireAuth, handler.ChangePassword)

	sessions := router.Group("/me", requireAuth)
	sessions.Get("/sessions", handler.Sessions)
	sessions.Delete("/sessions/:sessionID", handler.RevokeSession)
}
