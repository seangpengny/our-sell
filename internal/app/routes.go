package app

import (
	"context"
	"time"

	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"

	apiSpec "github.com/vtech/our-sell/api"
	"github.com/vtech/our-sell/internal/auth"
	"github.com/vtech/our-sell/internal/middleware"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/response"
)

func registerRoutes(app *fiber.App, deps Dependencies, authHandler *auth.Handler, tokenService *auth.TokenService) {
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(deps.Logger))
	app.Use(middleware.Recovery(deps.Logger))
	app.Use(middleware.SecurityHeaders(deps.Config.AppEnv == "production"))
	app.Use(middleware.CORS(deps.Config.CORSAllowedOrigins))
	app.Use(middleware.RateLimit(deps.Redis, deps.Config.RateLimitFailOpen))
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath:    "/",
		FilePath:    "openapi.yaml",
		FileContent: apiSpec.OpenAPISpec,
		Path:        "docs",
		Title:       "Our Sell API documentation",
		CacheAge:    300,
	}))
	app.Get("/docs/", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusMovedPermanently).To("/docs")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(response.Message{Message: "ok"})
	})
	app.Get("/ready", readiness(deps))

	api := app.Group("/api")
	v1 := api.Group("/v1")
	requireAuth := middleware.RequireAuth(tokenService)
	auth.RegisterRoutes(v1, authHandler, requireAuth)

	app.Use(func(c fiber.Ctx) error { return apperror.ErrNotFound })
}

func readiness(deps Dependencies) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if err := deps.Postgres.Ping(ctx); err != nil {
			return apperror.ErrDependencyUnavailable
		}
		if err := deps.Redis.Ping(ctx).Err(); err != nil {
			return apperror.ErrDependencyUnavailable
		}
		return c.JSON(response.Message{Message: "ready"})
	}
}
