package email

import (
	"context"
	"log/slog"
)

type Mailer interface {
	SendVerification(ctx context.Context, recipient, token string) error
	SendPasswordReset(ctx context.Context, recipient, token string) error
}

// LogMailer is deliberately token-blind. A real SMTP/provider adapter can
// implement Mailer without changing authentication or password-reset logic.
type LogMailer struct {
	logger *slog.Logger
}

func NewLogMailer(logger *slog.Logger) *LogMailer { return &LogMailer{logger: logger} }

func (m *LogMailer) SendVerification(ctx context.Context, recipient, _ string) error {
	m.logger.InfoContext(ctx, "verification email queued", "recipient", recipient)
	return nil
}

func (m *LogMailer) SendPasswordReset(ctx context.Context, recipient, _ string) error {
	m.logger.InfoContext(ctx, "password reset email queued", "recipient", recipient)
	return nil
}
