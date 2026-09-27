package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	content := `
# Comment line
APP_ENV=test-env
EXPORT_VAR=exported
export ANOTHER_VAR="quoted value"
SINGLE_QUOTED='single quoted'
ALREADY_SET=new_value
EMPTY_VAL=
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test .env file: %v", err)
	}

	t.Setenv("ALREADY_SET", "existing_value")

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	loadDotEnv()

	if got := os.Getenv("APP_ENV"); got != "test-env" {
		t.Errorf("expected APP_ENV=test-env, got %q", got)
	}
	if got := os.Getenv("EXPORT_VAR"); got != "exported" {
		t.Errorf("expected EXPORT_VAR=exported, got %q", got)
	}
	if got := os.Getenv("ANOTHER_VAR"); got != "quoted value" {
		t.Errorf("expected ANOTHER_VAR='quoted value', got %q", got)
	}
	if got := os.Getenv("SINGLE_QUOTED"); got != "single quoted" {
		t.Errorf("expected SINGLE_QUOTED='single quoted', got %q", got)
	}
	if got := os.Getenv("ALREADY_SET"); got != "existing_value" {
		t.Errorf("expected ALREADY_SET='existing_value', got %q", got)
	}
}

func TestLoadValidation(t *testing.T) {
	t.Run("missing database url", func(t *testing.T) {
		tempDir := t.TempDir()
		origWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer func() { _ = os.Chdir(origWd) }()

		t.Setenv("DATABASE_URL", "")
		_, err = Load()
		if err == nil || err.Error() != "DATABASE_URL is required" {
			t.Errorf("expected 'DATABASE_URL is required', got %v", err)
		}
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
		t.Setenv("REDIS_URL", "redis://localhost:6379/0")
		t.Setenv("JWT_SECRET", "super-secret-key-that-is-at-least-32-bytes-long")
		t.Setenv("JWT_ISSUER", "our-sell-api")
		t.Setenv("JWT_AUDIENCE", "our-sell-client")
		t.Setenv("ACCESS_TOKEN_TTL", "15m")
		t.Setenv("REFRESH_TOKEN_TTL", "720h")
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_USERNAME", "mailer@example.com")
		t.Setenv("SMTP_PASSWORD", "password")
		t.Setenv("SMTP_FROM", "no-reply@example.com")
		t.Setenv("SMTP_TLS_MODE", "starttls")
		t.Setenv("SMTP_TIMEOUT", "15s")
		t.Setenv("EMAIL_VERIFICATION_URL", "http://localhost:3000/verify")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/db?sslmode=disable" {
			t.Errorf("unexpected database URL: %s", cfg.DatabaseURL)
		}
		if cfg.AccessTokenTTL != 15*time.Minute {
			t.Errorf("expected 15m access token TTL, got %v", cfg.AccessTokenTTL)
		}
		if cfg.SMTPFrom != "no-reply@example.com" || cfg.SMTPTimeout != 15*time.Second {
			t.Errorf("unexpected SMTP config: from=%q timeout=%v", cfg.SMTPFrom, cfg.SMTPTimeout)
		}
	})

	t.Run("production requires SMTP", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("COOKIE_SECURE", "true")
		t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
		t.Setenv("REDIS_URL", "redis://localhost:6379/0")
		t.Setenv("JWT_SECRET", "super-secret-key-that-is-at-least-32-bytes-long")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SMTP_USERNAME", "")
		t.Setenv("SMTP_PASSWORD", "")
		t.Setenv("SMTP_FROM", "")

		_, err := Load()
		if err == nil || err.Error() != "SMTP_HOST is required in production" {
			t.Errorf("expected production SMTP error, got %v", err)
		}
	})
}
