package email

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSMTPMailerSendVerification(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	messages := make(chan string, 1)
	go serveTestSMTP(t, listener, messages)

	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:            "127.0.0.1",
		Port:            listener.Addr().(*net.TCPAddr).Port,
		From:            "no-reply@example.com",
		FromName:        "Our Sell",
		TLSMode:         "plain",
		Timeout:         2 * time.Second,
		VerificationURL: "https://app.example.com/verify?source=email",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("new mailer: %v", err)
	}

	token := strings.Repeat("a", 64)
	if err := mailer.SendVerification(context.Background(), "person@example.com", token); err != nil {
		t.Fatalf("send verification: %v", err)
	}

	select {
	case message := <-messages:
		if !strings.Contains(message, "Subject: Verify your Our Sell email") {
			t.Errorf("message subject missing: %s", message)
		}
		if !strings.Contains(message, token) {
			t.Errorf("message token missing")
		}
		if !strings.Contains(message, "source=3Demail") || !strings.Contains(message, "token=3D") {
			t.Errorf("verification URL missing: %s", message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SMTP message")
	}
}

func TestNewSMTPMailerValidation(t *testing.T) {
	tests := []struct {
		name string
		cfg  SMTPConfig
		want string
	}{
		{name: "missing host", cfg: SMTPConfig{From: "no-reply@example.com"}, want: ErrSMTPHostRequired.Error()},
		{name: "plain authentication", cfg: SMTPConfig{Host: "smtp.example.com", Username: "user@example.com", Password: "password", From: "no-reply@example.com", TLSMode: "plain"}, want: "SMTP_TLS_MODE=plain cannot be used with SMTP authentication"},
		{name: "invalid URL", cfg: SMTPConfig{Host: "smtp.example.com", From: "no-reply@example.com", VerificationURL: "javascript:alert(1)"}, want: "EMAIL_VERIFICATION_URL: URL must be an absolute HTTP(S) URL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewSMTPMailer(test.cfg, nil)
			if err == nil || err.Error() != test.want {
				t.Errorf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestTokenLink(t *testing.T) {
	if got := tokenLink("https://app.example.com/reset", "token-value"); got != "https://app.example.com/reset?token=token-value" {
		t.Errorf("unexpected token link: %q", got)
	}
	if got := tokenLink("https://app.example.com/reset/{token}", "token value"); got != "https://app.example.com/reset/token+value" {
		t.Errorf("unexpected placeholder token link: %q", got)
	}
}

func serveTestSMTP(t *testing.T, listener net.Listener, messages chan<- string) {
	t.Helper()
	connection, err := listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = connection.Close() }()
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	writeResponse := func(response string) {
		_, _ = fmt.Fprint(writer, response+"\r\n")
		_ = writer.Flush()
	}
	writeResponse("220 test SMTP server")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(command, "EHLO"), strings.HasPrefix(command, "HELO"):
			_, _ = fmt.Fprint(writer, "250-localhost\r\n250 PIPELINING\r\n")
			_ = writer.Flush()
		case strings.HasPrefix(command, "MAIL FROM:"), strings.HasPrefix(command, "RCPT TO:"):
			writeResponse("250 accepted")
		case command == "DATA":
			writeResponse("354 end with <CRLF>.<CRLF>")
			var message strings.Builder
			for {
				dataLine, readErr := reader.ReadString('\n')
				if readErr != nil {
					return
				}
				dataLine = strings.TrimSuffix(dataLine, "\n")
				dataLine = strings.TrimSuffix(dataLine, "\r")
				if dataLine == "." {
					break
				}
				message.WriteString(dataLine)
				message.WriteString("\n")
			}
			messages <- message.String()
			writeResponse("250 queued")
		case command == "QUIT":
			writeResponse("221 closing")
			return
		default:
			writeResponse("250 accepted")
		}
	}
}
