package config

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL string
	RedisURL    string

	DBMaxConns          int32
	DBMinConns          int32
	DBMaxConnLifetime   time.Duration
	DBMaxConnIdleTime   time.Duration
	DBHealthCheckPeriod time.Duration

	JWTSecret       string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	CORSAllowedOrigins    []string
	TrustedProxyCIDRs     []string
	TrustForwardedHeads   bool
	RequestBodyLimit      int
	ShutdownTimeout       time.Duration
	RefreshCookieName     string
	CookieDomain          string
	CookieSecure          bool
	CookieSameSite        string
	VerificationTokenTTL  time.Duration
	PasswordResetTokenTTL time.Duration

	RateLimitFailOpen bool
}

func Load() (Config, error) {
	loadDotEnv()

	c := Config{
		AppEnv:                envOr("APP_ENV", "development"),
		AppPort:               envOr("APP_PORT", "8080"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              os.Getenv("REDIS_URL"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		JWTIssuer:             envOr("JWT_ISSUER", "our-sell-api"),
		JWTAudience:           envOr("JWT_AUDIENCE", "our-sell-client"),
		AccessTokenTTL:        durationOr("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:       durationOr("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		DBMaxConns:            int32Or("DB_MAX_CONNS", 10),
		DBMinConns:            int32Or("DB_MIN_CONNS", 2),
		DBMaxConnLifetime:     durationOr("DB_MAX_CONN_LIFETIME", time.Hour),
		DBMaxConnIdleTime:     durationOr("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		DBHealthCheckPeriod:   durationOr("DB_HEALTH_CHECK_PERIOD", time.Minute),
		CORSAllowedOrigins:    listOr("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxyCIDRs:     listOr("TRUSTED_PROXY_CIDRS", nil),
		TrustForwardedHeads:   boolOr("TRUST_FORWARDED_HEADERS", false),
		RequestBodyLimit:      intOr("REQUEST_BODY_LIMIT", 1<<20),
		ShutdownTimeout:       durationOr("SHUTDOWN_TIMEOUT", 10*time.Second),
		RefreshCookieName:     envOr("REFRESH_COOKIE_NAME", "refresh_token"),
		CookieDomain:          os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:          boolOr("COOKIE_SECURE", false),
		CookieSameSite:        strings.ToLower(envOr("COOKIE_SAMESITE", "lax")),
		VerificationTokenTTL:  durationOr("VERIFICATION_TOKEN_TTL", 24*time.Hour),
		PasswordResetTokenTTL: durationOr("PASSWORD_RESET_TOKEN_TTL", time.Hour),
		RateLimitFailOpen:     boolOr("RATE_LIMIT_FAIL_OPEN", false),
	}

	if c.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 bytes")
	}
	if c.JWTIssuer == "" || c.JWTAudience == "" {
		return Config{}, errors.New("JWT_ISSUER and JWT_AUDIENCE are required")
	}
	if c.AccessTokenTTL < 10*time.Minute || c.AccessTokenTTL > 30*time.Minute {
		return Config{}, errors.New("ACCESS_TOKEN_TTL must be between 10m and 30m")
	}
	if c.RefreshTokenTTL <= c.AccessTokenTTL {
		return Config{}, errors.New("REFRESH_TOKEN_TTL must be longer than ACCESS_TOKEN_TTL")
	}
	if c.DBMaxConnLifetime <= 0 || c.DBMaxConnIdleTime <= 0 || c.DBHealthCheckPeriod <= 0 || c.ShutdownTimeout <= 0 || c.VerificationTokenTTL <= 0 || c.PasswordResetTokenTTL <= 0 {
		return Config{}, errors.New("duration configuration values must be positive")
	}
	if c.DBMaxConns < 1 || c.DBMinConns < 0 || c.DBMinConns > c.DBMaxConns {
		return Config{}, errors.New("DB pool connection limits are invalid")
	}
	if c.RequestBodyLimit < 1024 {
		return Config{}, errors.New("REQUEST_BODY_LIMIT must be at least 1024 bytes")
	}
	if c.CookieSameSite != "lax" && c.CookieSameSite != "strict" && c.CookieSameSite != "none" {
		return Config{}, errors.New("COOKIE_SAMESITE must be lax, strict, or none")
	}
	if c.CookieSameSite == "none" && !c.CookieSecure {
		return Config{}, errors.New("COOKIE_SECURE must be true when COOKIE_SAMESITE is none")
	}
	if c.AppEnv == "production" && !c.CookieSecure {
		return Config{}, errors.New("COOKIE_SECURE must be true in production")
	}
	if c.TrustForwardedHeads && len(c.TrustedProxyCIDRs) == 0 {
		return Config{}, errors.New("TRUSTED_PROXY_CIDRS is required when TRUST_FORWARDED_HEADERS is true")
	}
	for _, proxy := range c.TrustedProxyCIDRs {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return Config{}, fmt.Errorf("invalid trusted proxy address %q", proxy)
			}
		}
	}
	for _, origin := range c.CORSAllowedOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return Config{}, fmt.Errorf("invalid CORS origin %q", origin)
		}
		if origin == "*" {
			return Config{}, errors.New("CORS_ALLOWED_ORIGINS cannot contain *")
		}
	}
	return c, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func listOr(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func durationOr(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return parsed
}

func intOr(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func int32Or(key string, fallback int32) int32 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return int32(parsed)
}

func boolOr(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}

func loadDotEnv() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			key, val, found := strings.Cut(line, "=")
			if !found {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			if len(val) >= 2 {
				if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
					(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
					val = val[1 : len(val)-1]
				}
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
		_ = f.Close()
		return
	}
}

