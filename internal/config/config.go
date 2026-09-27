package config

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/mail"
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

	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
	SMTPFromName         string
	SMTPTLSMode          string
	SMTPTimeout          time.Duration
	EmailVerificationURL string
	PasswordResetURL     string

	FacebookAppID              string
	FacebookAppSecret          string
	FacebookRedirectURI        string
	FacebookFrontendURL        string
	MetaGraphAPIVersion        string
	FacebookTokenEncryptionKey []byte

	BakongAPIToken      string
	BakongAPIBaseURL    string
	BakongAccountID     string
	BakongMerchantName  string
	BakongMerchantCity  string
	BakongMerchantID    string
	BakongAcquiringBank string
	BakongStoreLabel    string
	BakongTopupTTL      time.Duration
	BakongMinTopupUSD   float64
	BakongMaxTopupUSD   float64
	BakongWebhookSecret string
	BakongHTTPTimeout   time.Duration

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
		SMTPHost:              strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:              intOr("SMTP_PORT", 587),
		SMTPUsername:          strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:          os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:              strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPFromName:          strings.TrimSpace(envOr("SMTP_FROM_NAME", "Our Sell")),
		SMTPTLSMode:           strings.ToLower(envOr("SMTP_TLS_MODE", "starttls")),
		SMTPTimeout:           durationOr("SMTP_TIMEOUT", 10*time.Second),
		EmailVerificationURL:  strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_URL")),
		PasswordResetURL:      strings.TrimSpace(os.Getenv("PASSWORD_RESET_URL")),
		FacebookAppID:         strings.TrimSpace(os.Getenv("FACEBOOK_APP_ID")),
		FacebookAppSecret:     os.Getenv("FACEBOOK_APP_SECRET"),
		FacebookRedirectURI:   strings.TrimSpace(os.Getenv("FACEBOOK_REDIRECT_URI")),
		FacebookFrontendURL:   strings.TrimSpace(envOr("FACEBOOK_FRONTEND_URL", "http://localhost:3000")),
		MetaGraphAPIVersion:   strings.TrimSpace(envOr("META_GRAPH_API_VERSION", "v26.0")),
		BakongAPIToken:        strings.TrimSpace(os.Getenv("BAKONG_API_TOKEN")),
		BakongAPIBaseURL:      strings.TrimRight(strings.TrimSpace(envOr("BAKONG_API_BASE_URL", "https://api-bakong.nbc.gov.kh")), "/"),
		BakongAccountID:       strings.TrimSpace(os.Getenv("BAKONG_ACCOUNT_ID")),
		BakongMerchantName:    strings.TrimSpace(envOr("BAKONG_MERCHANT_NAME", "Our Sell")),
		BakongMerchantCity:    strings.TrimSpace(envOr("BAKONG_MERCHANT_CITY", "Phnom Penh")),
		BakongMerchantID:      strings.TrimSpace(os.Getenv("BAKONG_MERCHANT_ID")),
		BakongAcquiringBank:   strings.TrimSpace(os.Getenv("BAKONG_ACQUIRING_BANK")),
		BakongStoreLabel:      strings.TrimSpace(envOr("BAKONG_STORE_LABEL", "Our Sell Wallet")),
		BakongTopupTTL:        durationOr("BAKONG_TOPUP_TTL", 15*time.Minute),
		BakongMinTopupUSD:     float64Or("BAKONG_MIN_TOPUP_USD", 1),
		BakongMaxTopupUSD:     float64Or("BAKONG_MAX_TOPUP_USD", 10000),
		BakongWebhookSecret:   os.Getenv("BAKONG_WEBHOOK_SECRET"),
		BakongHTTPTimeout:     durationOr("BAKONG_HTTP_TIMEOUT", 10*time.Second),
		RateLimitFailOpen:     boolOr("RATE_LIMIT_FAIL_OPEN", false),
	}
	if encodedKey := strings.TrimSpace(os.Getenv("FACEBOOK_TOKEN_ENCRYPTION_KEY")); encodedKey != "" {
		decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
		if err != nil {
			return Config{}, errors.New("FACEBOOK_TOKEN_ENCRYPTION_KEY must be valid base64")
		}
		c.FacebookTokenEncryptionKey = decodedKey
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
	if err := validateSMTP(&c); err != nil {
		return Config{}, err
	}
	if err := validateOptionalHTTPURL("EMAIL_VERIFICATION_URL", c.EmailVerificationURL); err != nil {
		return Config{}, err
	}
	if err := validateOptionalHTTPURL("PASSWORD_RESET_URL", c.PasswordResetURL); err != nil {
		return Config{}, err
	}
	if err := validateFacebook(&c); err != nil {
		return Config{}, err
	}
	if err := validateBakong(&c); err != nil {
		return Config{}, err
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

func validateBakong(c *Config) error {
	configured := c.BakongAPIToken != "" || c.BakongAccountID != ""
	if !configured {
		return nil
	}
	if c.BakongAPIToken == "" || c.BakongAccountID == "" {
		return errors.New("BAKONG_API_TOKEN and BAKONG_ACCOUNT_ID are required together")
	}
	if c.BakongMerchantName == "" || c.BakongMerchantCity == "" {
		return errors.New("BAKONG_MERCHANT_NAME and BAKONG_MERCHANT_CITY are required when Bakong is configured")
	}
	if err := validateOptionalHTTPURL("BAKONG_API_BASE_URL", c.BakongAPIBaseURL); err != nil {
		return err
	}
	if c.BakongTopupTTL < time.Minute || c.BakongTopupTTL > 24*time.Hour {
		return errors.New("BAKONG_TOPUP_TTL must be between 1m and 24h")
	}
	if c.BakongHTTPTimeout <= 0 || c.BakongHTTPTimeout > time.Minute {
		return errors.New("BAKONG_HTTP_TIMEOUT must be between 1s and 1m")
	}
	if c.BakongMinTopupUSD <= 0 || c.BakongMaxTopupUSD < c.BakongMinTopupUSD {
		return errors.New("Bakong top-up limits are invalid")
	}
	return nil
}

func validateFacebook(c *Config) error {
	configured := c.FacebookAppID != "" || c.FacebookAppSecret != "" || len(c.FacebookTokenEncryptionKey) > 0
	if !configured {
		return nil
	}
	if c.FacebookAppID == "" || c.FacebookAppSecret == "" || c.FacebookRedirectURI == "" || len(c.FacebookTokenEncryptionKey) == 0 {
		return errors.New("FACEBOOK_APP_ID, FACEBOOK_APP_SECRET, FACEBOOK_REDIRECT_URI, and FACEBOOK_TOKEN_ENCRYPTION_KEY are required together")
	}
	if len(c.FacebookTokenEncryptionKey) != 32 {
		return errors.New("FACEBOOK_TOKEN_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	if err := validateOptionalHTTPURL("FACEBOOK_REDIRECT_URI", c.FacebookRedirectURI); err != nil {
		return err
	}
	if err := validateOptionalHTTPURL("FACEBOOK_FRONTEND_URL", c.FacebookFrontendURL); err != nil {
		return err
	}
	version := strings.TrimPrefix(c.MetaGraphAPIVersion, "v")
	versionParts := strings.Split(version, ".")
	major, majorErr := strconv.Atoi(versionParts[0])
	minor, minorErr := 0, error(nil)
	if len(versionParts) == 2 {
		minor, minorErr = strconv.Atoi(versionParts[1])
	}
	if c.MetaGraphAPIVersion == "" || !strings.HasPrefix(c.MetaGraphAPIVersion, "v") || len(versionParts) != 2 || major < 1 || minor < 0 || majorErr != nil || minorErr != nil {
		return errors.New("META_GRAPH_API_VERSION must look like v26.0")
	}
	if c.AppEnv == "production" {
		redirect, _ := url.Parse(c.FacebookRedirectURI)
		frontend, _ := url.Parse(c.FacebookFrontendURL)
		if redirect.Scheme != "https" || frontend.Scheme != "https" {
			return errors.New("facebook redirect and frontend URLs must use HTTPS in production")
		}
	}
	return nil
}

func validateSMTP(c *Config) error {
	smtpConfigured := c.SMTPHost != "" || c.SMTPUsername != "" || c.SMTPPassword != "" || c.SMTPFrom != ""
	if c.AppEnv == "production" && c.SMTPHost == "" {
		return errors.New("SMTP_HOST is required in production")
	}
	if !smtpConfigured {
		return nil
	}
	if c.SMTPHost == "" {
		return errors.New("SMTP_HOST is required when SMTP settings are configured")
	}
	if strings.ContainsAny(c.SMTPHost, "\r\n /@") {
		return errors.New("SMTP_HOST contains invalid characters")
	}
	if c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return errors.New("SMTP_PORT must be between 1 and 65535")
	}
	if c.SMTPUsername == "" && c.SMTPPassword != "" {
		return errors.New("SMTP_USERNAME is required when SMTP_PASSWORD is set")
	}
	if c.SMTPUsername != "" && c.SMTPPassword == "" {
		return errors.New("SMTP_PASSWORD is required when SMTP_USERNAME is set")
	}
	if c.SMTPFrom == "" {
		c.SMTPFrom = c.SMTPUsername
	}
	if c.SMTPFrom == "" {
		return errors.New("SMTP_FROM is required when SMTP_USERNAME is empty")
	}
	if strings.ContainsAny(c.SMTPFromName, "\r\n") {
		return errors.New("SMTP_FROM_NAME contains invalid characters")
	}
	if _, err := mail.ParseAddress(c.SMTPFrom); err != nil {
		return fmt.Errorf("SMTP_FROM must be a valid email address: %w", err)
	}
	if c.SMTPTLSMode != "starttls" && c.SMTPTLSMode != "tls" && c.SMTPTLSMode != "plain" {
		return errors.New("SMTP_TLS_MODE must be starttls, tls, or plain")
	}
	if c.SMTPTLSMode == "plain" && c.AppEnv == "production" {
		return errors.New("SMTP_TLS_MODE=plain is not allowed in production")
	}
	if c.SMTPTLSMode == "plain" && c.SMTPUsername != "" {
		return errors.New("SMTP_TLS_MODE=plain cannot be used with SMTP authentication")
	}
	if c.SMTPTimeout <= 0 || c.SMTPTimeout > 2*time.Minute {
		return errors.New("SMTP_TIMEOUT must be between 1s and 2m")
	}
	return nil
}

func validateOptionalHTTPURL(name, value string) error {
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s contains invalid characters", name)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be an absolute HTTP(S) URL", name)
	}
	return nil
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

func float64Or(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
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
