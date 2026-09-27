package apperror

import (
	"errors"
	"fmt"
)

const (
	CodeValidation            = "VALIDATION_ERROR"
	CodeInvalidCredentials    = "INVALID_CREDENTIALS"
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeForbidden             = "FORBIDDEN"
	CodeNotFound              = "NOT_FOUND"
	CodeConflict              = "CONFLICT"
	CodeRateLimited           = "RATE_LIMITED"
	CodeDependencyUnavailable = "DEPENDENCY_UNAVAILABLE"
	CodeInsufficientFunds     = "INSUFFICIENT_FUNDS"
	CodeInternal              = "INTERNAL_ERROR"
)

var (
	ErrInvalidCredentials    = &Error{Status: 401, Code: CodeInvalidCredentials, Message: "Invalid email or password"}
	ErrUnauthorized          = &Error{Status: 401, Code: CodeUnauthorized, Message: "Authentication required"}
	ErrForbidden             = &Error{Status: 403, Code: CodeForbidden, Message: "You do not have permission to perform this action"}
	ErrNotFound              = &Error{Status: 404, Code: CodeNotFound, Message: "Resource not found"}
	ErrConflict              = &Error{Status: 409, Code: CodeConflict, Message: "Resource already exists"}
	ErrRateLimited           = &Error{Status: 429, Code: CodeRateLimited, Message: "Too many requests"}
	ErrInvalidToken          = &Error{Status: 401, Code: CodeUnauthorized, Message: "Invalid or expired token"}
	ErrDependencyUnavailable = &Error{Status: 503, Code: CodeDependencyUnavailable, Message: "A required service is temporarily unavailable"}
	ErrInsufficientFunds     = &Error{Status: 402, Code: CodeInsufficientFunds, Message: "Your wallet balance is not sufficient for this purchase"}
)

type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

func NewValidation(fields map[string]string) *Error {
	return &Error{Status: 400, Code: CodeValidation, Message: "Invalid request", Fields: fields}
}

func Wrap(err error, message string) error {
	return &Error{Status: 500, Code: CodeInternal, Message: message, Err: err}
}

func Is(err, target error) bool { return errors.Is(err, target) }

func Public(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return &Error{Status: 500, Code: CodeInternal, Message: "Internal server error", Err: err}
}
