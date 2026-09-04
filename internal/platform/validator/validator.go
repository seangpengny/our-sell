package validator

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func NormalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if len(email) == 0 || len(email) > 320 || !emailPattern.MatchString(email) {
		return "", errors.New("invalid email address")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", errors.New("invalid email address")
	}
	return email, nil
}

func Password(value string) error {
	if len(value) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if len(value) > 128 {
		return errors.New("password must be at most 128 characters")
	}
	return nil
}

func Name(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" || len(name) > 100 {
		return "", errors.New("name must be between 1 and 100 characters")
	}
	return name, nil
}

func UUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID")
	}
	return parsed, nil
}
