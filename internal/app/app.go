package app

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/vtech/our-sell/db/sqlc"
	"github.com/vtech/our-sell/internal/auth"
	"github.com/vtech/our-sell/internal/config"
	"github.com/vtech/our-sell/internal/email"
	"github.com/vtech/our-sell/internal/facebook"
	"github.com/vtech/our-sell/internal/middleware"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/response"
	"github.com/vtech/our-sell/internal/realtime"
	"github.com/vtech/our-sell/internal/user"
	"github.com/vtech/our-sell/internal/wallet"
)

type Dependencies struct {
	Config   config.Config
	Logger   *slog.Logger
	Postgres *pgxpool.Pool
	Redis    *redis.Client
}

type Application struct {
	Fiber        *fiber.App
	AuthService  *auth.Service
	UserService  *user.Service
	RealtimeHub  *realtime.Hub
	TokenService *auth.TokenService
}

func New(deps Dependencies) (*Application, error) {
	queries := sqlc.New(deps.Postgres)
	userRepository := user.NewRepository(queries)
	userService := user.NewService(userRepository)
	authRepository := auth.NewRepository(queries, deps.Postgres)
	tokenService, err := auth.NewTokenService(deps.Config.JWTSecret, deps.Config.JWTIssuer, deps.Config.JWTAudience, deps.Config.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	mailer, err := email.NewMailer(email.SMTPConfig{
		Host:             deps.Config.SMTPHost,
		Port:             deps.Config.SMTPPort,
		Username:         deps.Config.SMTPUsername,
		Password:         deps.Config.SMTPPassword,
		From:             deps.Config.SMTPFrom,
		FromName:         deps.Config.SMTPFromName,
		TLSMode:          deps.Config.SMTPTLSMode,
		Timeout:          deps.Config.SMTPTimeout,
		VerificationURL:  deps.Config.EmailVerificationURL,
		PasswordResetURL: deps.Config.PasswordResetURL,
	}, deps.Logger)
	if err != nil {
		return nil, err
	}
	authService := auth.NewService(userRepository, authRepository, mailer, tokenService, deps.Config.VerificationTokenTTL, deps.Config.PasswordResetTokenTTL, deps.Config.RefreshTokenTTL)
	authHandler := auth.NewHandler(authService, auth.CookieConfig{
		Name: deps.Config.RefreshCookieName, Domain: deps.Config.CookieDomain, Secure: deps.Config.CookieSecure,
		SameSite: deps.Config.CookieSameSite, MaxAge: int(deps.Config.RefreshTokenTTL.Seconds()),
	})
	userHandler := user.NewHandler(userService)
	var facebookClient *facebook.Client
	if deps.Config.FacebookAppID != "" {
		facebookClient, err = facebook.NewClient(facebook.ClientConfig{
			AppID: deps.Config.FacebookAppID, AppSecret: deps.Config.FacebookAppSecret,
			RedirectURI: deps.Config.FacebookRedirectURI, Version: deps.Config.MetaGraphAPIVersion,
		})
		if err != nil {
			return nil, err
		}
	}
	var facebookCipher *facebook.TokenCipher
	if len(deps.Config.FacebookTokenEncryptionKey) > 0 {
		facebookCipher, err = facebook.NewTokenCipher(deps.Config.FacebookTokenEncryptionKey)
		if err != nil {
			return nil, err
		}
	}
	facebookRepository := facebook.NewRepository(queries)
	facebookService := facebook.NewService(facebookRepository, facebookClient, facebookCipher, facebook.NewRedisStateStore(deps.Redis), deps.Config.FacebookFrontendURL, deps.Logger)
	facebookHandler := facebook.NewHandler(facebookService)
	walletRepository := wallet.NewRepository(queries, deps.Postgres)
	walletService := wallet.NewService(walletRepository, wallet.NewProvider(deps.Config), deps.Config, deps.Logger)
	walletHandler := wallet.NewHandler(walletService, deps.Config.BakongWebhookSecret)

	app := fiber.New(fiber.Config{
		AppName:          "our-sell-api",
		BodyLimit:        deps.Config.RequestBodyLimit,
		ServerHeader:     "",
		Immutable:        true,
		ReadTimeout:      15 * time.Second,
		WriteTimeout:     15 * time.Second,
		IdleTimeout:      60 * time.Second,
		TrustProxy:       deps.Config.TrustForwardedHeads,
		ProxyHeader:      "X-Forwarded-For",
		TrustProxyConfig: fiber.TrustProxyConfig{Proxies: deps.Config.TrustedProxyCIDRs},
		ErrorHandler:     errorHandler(deps.Logger),
	})

	hub := realtime.NewHub()
	registerRoutes(app, deps, authHandler, userHandler, facebookHandler, walletHandler, tokenService)
	return &Application{Fiber: app, AuthService: authService, UserService: userService, RealtimeHub: hub, TokenService: tokenService}, nil
}

func errorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		publicErr := apperror.Public(err)
		status := publicErr.Status
		code := publicErr.Code
		message := publicErr.Message
		fields := publicErr.Fields
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
			switch status {
			case http.StatusForbidden:
				code, message = apperror.CodeForbidden, "You do not have permission to perform this action"
			case http.StatusNotFound:
				code, message = apperror.CodeNotFound, "Resource not found"
			default:
				code, message = "HTTP_ERROR", http.StatusText(status)
			}
		}
		if status < 400 || status > 599 {
			status = http.StatusInternalServerError
		}
		if status >= 500 {
			logger.ErrorContext(c.Context(), "request error", "request_id", middleware.RequestIDFrom(c), "error", err)
		}
		return c.Status(status).JSON(response.ErrorBody{Error: response.ErrorDetails{Code: code, Message: message, Fields: fields}})
	}
}
