package errors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code       string
	Message    string
	Details    map[string]interface{}
	Underlying error
}

func (e *AppError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Underlying)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Underlying
}

func NewAppError(code, message string, underlying error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Underlying: underlying,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal server error")
	ErrBadRequest   = errors.New("bad request")
	ErrRateLimit    = errors.New("rate limit exceeded")
	ErrTimeout      = errors.New("request timeout")
	ErrCircuitOpen  = errors.New("circuit breaker open")
)

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
