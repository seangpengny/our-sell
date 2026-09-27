package email

import (
	"context"
	"log/slog"
	"strings"
)

type Mailer interface {
	SendVerification(ctx context.Context, recipient, token string) error
	SendPasswordReset(ctx context.Context, recipient, token string) error
}

// NewMailer selects SMTP delivery when an SMTP host is configured. With no
// SMTP host, development keeps using a token-blind logger mailer.
func NewMailer(cfg SMTPConfig, logger *slog.Logger) (Mailer, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		if cfg.Username != "" || cfg.Password != "" || cfg.From != "" {
			return nil, ErrSMTPHostRequired
		}
		return NewLogMailer(logger), nil
	}
	return NewSMTPMailer(cfg, logger)
}

// LogMailer is deliberately token-blind and is only intended for local
// development when SMTP delivery is not configured.
type LogMailer struct {
	logger *slog.Logger
}

func NewLogMailer(logger *slog.Logger) *LogMailer {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogMailer{logger: logger}
}

func (m *LogMailer) SendVerification(ctx context.Context, recipient, _ string) error {
	m.logger.InfoContext(ctx, "verification email queued", "recipient", recipient)
	return nil
}

func (m *LogMailer) SendPasswordReset(ctx context.Context, recipient, _ string) error {
	m.logger.InfoContext(ctx, "password reset email queued", "recipient", recipient)
	return nil
}
