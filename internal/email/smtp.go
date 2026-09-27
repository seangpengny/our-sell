package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrSMTPHostRequired = errors.New("SMTP_HOST is required")

type SMTPConfig struct {
	Host             string
	Port             int
	Username         string
	Password         string
	From             string
	FromName         string
	TLSMode          string
	Timeout          time.Duration
	VerificationURL  string
	PasswordResetURL string
}

type SMTPMailer struct {
	cfg    SMTPConfig
	logger *slog.Logger
}

func NewSMTPMailer(cfg SMTPConfig, logger *slog.Logger) (*SMTPMailer, error) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.From = strings.TrimSpace(cfg.From)
	cfg.FromName = strings.TrimSpace(cfg.FromName)
	cfg.TLSMode = strings.ToLower(strings.TrimSpace(cfg.TLSMode))
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.TLSMode == "" {
		cfg.TLSMode = "starttls"
	}
	if cfg.Host == "" {
		return nil, ErrSMTPHostRequired
	}
	if strings.ContainsAny(cfg.Host, "\r\n /@") {
		return nil, errors.New("SMTP_HOST contains invalid characters")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("SMTP_PORT must be between 1 and 65535")
	}
	if cfg.Username == "" && cfg.Password != "" {
		return nil, errors.New("SMTP_USERNAME is required when SMTP_PASSWORD is set")
	}
	if cfg.Username != "" && cfg.Password == "" {
		return nil, errors.New("SMTP_PASSWORD is required when SMTP_USERNAME is set")
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	if cfg.From == "" {
		return nil, errors.New("SMTP_FROM is required when SMTP_USERNAME is empty")
	}
	if strings.ContainsAny(cfg.FromName, "\r\n") {
		return nil, errors.New("SMTP_FROM_NAME contains invalid characters")
	}
	from, err := mail.ParseAddress(cfg.From)
	if err != nil || from.Address == "" {
		return nil, fmt.Errorf("SMTP_FROM must be a valid email address: %w", err)
	}
	cfg.From = from.Address
	if cfg.FromName == "" {
		cfg.FromName = from.Name
	}
	if cfg.TLSMode != "starttls" && cfg.TLSMode != "tls" && cfg.TLSMode != "plain" {
		return nil, errors.New("SMTP_TLS_MODE must be starttls, tls, or plain")
	}
	if cfg.TLSMode == "plain" && cfg.Username != "" {
		return nil, errors.New("SMTP_TLS_MODE=plain cannot be used with SMTP authentication")
	}
	if cfg.Timeout <= 0 || cfg.Timeout > 2*time.Minute {
		return nil, errors.New("SMTP_TIMEOUT must be between 1s and 2m")
	}
	if err := validateURL(cfg.VerificationURL); err != nil {
		return nil, fmt.Errorf("EMAIL_VERIFICATION_URL: %w", err)
	}
	if err := validateURL(cfg.PasswordResetURL); err != nil {
		return nil, fmt.Errorf("PASSWORD_RESET_URL: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SMTPMailer{cfg: cfg, logger: logger}, nil
}

func (m *SMTPMailer) SendVerification(ctx context.Context, recipient, token string) error {
	if err := validateToken(token); err != nil {
		return err
	}
	link := tokenLink(m.cfg.VerificationURL, token)
	plain, htmlBody := tokenBodies("verify your email address", token, link)
	return m.send(ctx, "verification", recipient, "Verify your Our Sell email", plain, htmlBody)
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, recipient, token string) error {
	if err := validateToken(token); err != nil {
		return err
	}
	link := tokenLink(m.cfg.PasswordResetURL, token)
	plain, htmlBody := tokenBodies("reset your password", token, link)
	return m.send(ctx, "password reset", recipient, "Reset your Our Sell password", plain, htmlBody)
}

func (m *SMTPMailer) send(ctx context.Context, kind, recipient, subject, plainBody, htmlBody string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	recipientAddress, err := recipientAddress(recipient)
	if err != nil {
		return err
	}
	message, err := buildMessage(m.cfg.From, m.cfg.FromName, recipientAddress, subject, plainBody, htmlBody, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("build %s email: %w", kind, err)
	}
	sendCtx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()
	if err := sendCtx.Err(); err != nil {
		return err
	}

	conn, client, err := m.connect(sendCtx)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if err := client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(recipientAddress); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("close SMTP session: %w", err)
	}
	m.logger.InfoContext(ctx, "email sent", "kind", kind, "recipient", recipientAddress)
	return nil
}

func (m *SMTPMailer) connect(ctx context.Context) (net.Conn, *smtp.Client, error) {
	address := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	dialer := &net.Dialer{}
	var (
		conn net.Conn
		err  error
	)
	if m.cfg.TLSMode == "tls" {
		conn, err = (&tls.Dialer{
			NetDialer: dialer,
			Config:    &tls.Config{MinVersion: tls.VersionTLS12, ServerName: m.cfg.Host},
		}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return nil, nil, err
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(m.cfg.Timeout)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if err := client.Hello("our-sell-api"); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	if m.cfg.TLSMode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			_ = client.Close()
			return nil, nil, errors.New("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{MinVersion: tls.VersionTLS12, ServerName: m.cfg.Host}); err != nil {
			_ = client.Close()
			return nil, nil, err
		}
	}
	if m.cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)); err != nil {
			_ = client.Close()
			return nil, nil, err
		}
	}
	return conn, client, nil
}

func tokenBodies(action, token, link string) (string, string) {
	if link == "" {
		plain := fmt.Sprintf("Hello,\n\nUse the following token to %s:\n\n%s\n\nIf you did not request this email, you can ignore it.\n", action, token)
		htmlBody := fmt.Sprintf("<p>Hello,</p><p>Use the following token to %s:</p><p><code>%s</code></p><p>If you did not request this email, you can ignore it.</p>", action, html.EscapeString(token))
		return plain, htmlBody
	}
	plain := fmt.Sprintf("Hello,\n\nUse this link to %s:\n%s\n\nIf the link does not work, use this token:\n%s\n\nIf you did not request this email, you can ignore it.\n", action, link, token)
	htmlBody := fmt.Sprintf("<p>Hello,</p><p><a href=\"%s\">Click here to %s</a>.</p><p>If needed, use this token: <code>%s</code></p><p>If you did not request this email, you can ignore it.</p>", html.EscapeString(link), action, html.EscapeString(token))
	return plain, htmlBody
}

func tokenLink(rawURL, token string) string {
	if rawURL == "" {
		return ""
	}
	if strings.Contains(rawURL, "{token}") {
		return strings.ReplaceAll(rawURL, "{token}", url.QueryEscape(token))
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func buildMessage(from, fromName, recipient, subject, plainBody, htmlBody string, now time.Time) ([]byte, error) {
	for field, value := range map[string]string{"from": from, "from name": fromName, "recipient": recipient, "subject": subject} {
		if strings.ContainsAny(value, "\r\n") {
			return nil, fmt.Errorf("%s contains invalid line breaks", field)
		}
	}
	fromHeader := (&mail.Address{Name: fromName, Address: from}).String()
	toHeader := (&mail.Address{Address: recipient}).String()
	var message bytes.Buffer
	fmt.Fprintf(&message, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&message, "From: %s\r\n", fromHeader)
	fmt.Fprintf(&message, "To: %s\r\n", toHeader)
	fmt.Fprintf(&message, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	fmt.Fprint(&message, "MIME-Version: 1.0\r\n")
	writer := multipart.NewWriter(&message)
	fmt.Fprintf(&message, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", writer.Boundary())
	if err := writePart(writer, "text/plain; charset=UTF-8", plainBody); err != nil {
		return nil, err
	}
	if err := writePart(writer, "text/html; charset=UTF-8", htmlBody); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return message.Bytes(), nil
}

func writePart(writer *multipart.Writer, contentType, body string) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", contentType)
	header.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	encoder := quotedprintable.NewWriter(part)
	if _, err := io.WriteString(encoder, normalizeLineEndings(body)); err != nil {
		return err
	}
	return encoder.Close()
}

func normalizeLineEndings(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}

func recipientAddress(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", errors.New("recipient contains invalid line breaks")
	}
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil || parsed.Address == "" {
		return "", errors.New("recipient must be a valid email address")
	}
	return parsed.Address, nil
}

func validateToken(value string) error {
	if strings.TrimSpace(value) == "" || len(value) > 256 || strings.ContainsAny(value, "\r\n") {
		return errors.New("email token is invalid")
	}
	return nil
}

func validateURL(value string) error {
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("URL contains invalid line breaks")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("URL must be an absolute HTTP(S) URL")
	}
	return nil
}
