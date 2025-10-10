package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewWhatsAppError(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := NewWhatsAppError(
		ErrCodeSessionNotFound,
		"Session not found",
		http.StatusNotFound,
		underlyingErr,
	)

	if err.Code != ErrCodeSessionNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeSessionNotFound, err.Code)
	}

	if err.Message != "Session not found" {
		t.Errorf("Expected message 'Session not found', got %s", err.Message)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}

	if err.Underlying != underlyingErr {
		t.Errorf("Expected underlying error to be set")
	}
}

func TestWhatsAppError_Error(t *testing.T) {
	tests := []struct {
		name       string
		err        *WhatsAppError
		wantString string
	}{
		{
			name: "error without underlying",
			err: &WhatsAppError{
				Code:    ErrCodeSessionNotFound,
				Message: "Session not found",
			},
			wantString: "Session not found",
		},
		{
			name: "error with underlying",
			err: &WhatsAppError{
				Code:       ErrCodeSessionNotFound,
				Message:    "Session not found",
				Underlying: errors.New("database error"),
			},
			wantString: "Session not found: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantString {
				t.Errorf("Error() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

func TestWhatsAppError_WithDetails(t *testing.T) {
	err := NewWhatsAppError(
		ErrCodeSessionNotFound,
		"Session not found",
		http.StatusNotFound,
		nil,
	)

	details := map[string]interface{}{
		"session_id": "test-session",
		"user_id":    "test-user",
	}

	err.WithDetails(details)

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}

	if err.Details["session_id"] != "test-session" {
		t.Errorf("Expected session_id to be 'test-session', got %v", err.Details["session_id"])
	}

	if err.Details["user_id"] != "test-user" {
		t.Errorf("Expected user_id to be 'test-user', got %v", err.Details["user_id"])
	}
}

func TestWhatsAppError_WithDetail(t *testing.T) {
	err := NewWhatsAppError(
		ErrCodeSessionNotFound,
		"Session not found",
		http.StatusNotFound,
		nil,
	)

	err.WithDetail("session_id", "test-session")

	if len(err.Details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(err.Details))
	}

	if err.Details["session_id"] != "test-session" {
		t.Errorf("Expected session_id to be 'test-session', got %v", err.Details["session_id"])
	}
}

func TestNewSessionNotFoundError(t *testing.T) {
	sessionID := "test-session-123"
	err := NewSessionNotFoundError(sessionID)

	if err.Code != ErrCodeSessionNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeSessionNotFound, err.Code)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}

	if err.Details["session_id"] != sessionID {
		t.Errorf("Expected session_id detail to be %s, got %v", sessionID, err.Details["session_id"])
	}
}

func TestNewSessionNotConnectedError(t *testing.T) {
	sessionID := "test-session-123"
	err := NewSessionNotConnectedError(sessionID)

	if err.Code != ErrCodeSessionNotConnected {
		t.Errorf("Expected code %s, got %s", ErrCodeSessionNotConnected, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}

	if err.Details["session_id"] != sessionID {
		t.Errorf("Expected session_id detail to be %s, got %v", sessionID, err.Details["session_id"])
	}
}

func TestNewSessionAlreadyExistsError(t *testing.T) {
	userID := "user-123"
	sessionName := "my-session"
	err := NewSessionAlreadyExistsError(userID, sessionName)

	if err.Code != ErrCodeSessionAlreadyExists {
		t.Errorf("Expected code %s, got %s", ErrCodeSessionAlreadyExists, err.Code)
	}

	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusConflict, err.HTTPStatus)
	}

	if err.Details["user_id"] != userID {
		t.Errorf("Expected user_id detail to be %s, got %v", userID, err.Details["user_id"])
	}

	if err.Details["session_name"] != sessionName {
		t.Errorf("Expected session_name detail to be %s, got %v", sessionName, err.Details["session_name"])
	}
}

func TestNewUnauthorizedSessionError(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	err := NewUnauthorizedSessionError(userID, sessionID)

	if err.Code != ErrCodeUnauthorizedSession {
		t.Errorf("Expected code %s, got %s", ErrCodeUnauthorizedSession, err.Code)
	}

	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusForbidden, err.HTTPStatus)
	}
}

func TestNewQRCodeExpiredError(t *testing.T) {
	sessionID := "test-session-123"
	err := NewQRCodeExpiredError(sessionID)

	if err.Code != ErrCodeQRCodeExpired {
		t.Errorf("Expected code %s, got %s", ErrCodeQRCodeExpired, err.Code)
	}

	if err.HTTPStatus != http.StatusGone {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusGone, err.HTTPStatus)
	}
}

func TestNewMessageSendFailedError(t *testing.T) {
	sessionID := "session-123"
	recipient := "+1234567890"
	underlyingErr := errors.New("network error")
	err := NewMessageSendFailedError(sessionID, recipient, underlyingErr)

	if err.Code != ErrCodeMessageSendFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeMessageSendFailed, err.Code)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}

	if err.Underlying != underlyingErr {
		t.Errorf("Expected underlying error to be set")
	}
}

func TestNewInvalidPhoneNumberError(t *testing.T) {
	phoneNumber := "invalid-phone"
	err := NewInvalidPhoneNumberError(phoneNumber)

	if err.Code != ErrCodeInvalidPhoneNumber {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidPhoneNumber, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}

	if err.Details["phone_number"] != phoneNumber {
		t.Errorf("Expected phone_number detail to be %s, got %v", phoneNumber, err.Details["phone_number"])
	}
}

func TestNewMediaTooLargeError(t *testing.T) {
	size := int64(100 * 1024 * 1024)      // 100 MB
	maxSize := int64(50 * 1024 * 1024)    // 50 MB
	err := NewMediaTooLargeError(size, maxSize)

	if err.Code != ErrCodeMediaTooLarge {
		t.Errorf("Expected code %s, got %s", ErrCodeMediaTooLarge, err.Code)
	}

	if err.HTTPStatus != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusRequestEntityTooLarge, err.HTTPStatus)
	}

	if err.Details["size"] != size {
		t.Errorf("Expected size detail to be %d, got %v", size, err.Details["size"])
	}

	if err.Details["max_size"] != maxSize {
		t.Errorf("Expected max_size detail to be %d, got %v", maxSize, err.Details["max_size"])
	}
}

func TestNewRateLimitExceededError(t *testing.T) {
	retryAfter := 60
	err := NewRateLimitExceededError(retryAfter)

	if err.Code != ErrCodeRateLimitExceeded {
		t.Errorf("Expected code %s, got %s", ErrCodeRateLimitExceeded, err.Code)
	}

	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusTooManyRequests, err.HTTPStatus)
	}

	if err.Details["retry_after_seconds"] != retryAfter {
		t.Errorf("Expected retry_after_seconds detail to be %d, got %v", retryAfter, err.Details["retry_after_seconds"])
	}
}

func TestIsWhatsAppError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "WhatsApp error",
			err:  NewSessionNotFoundError("test"),
			want: true,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWhatsAppError(tt.err); got != tt.want {
				t.Errorf("IsWhatsAppError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWhatsAppError(t *testing.T) {
	waErr := NewSessionNotFoundError("test")
	stdErr := errors.New("standard error")

	tests := []struct {
		name    string
		err     error
		wantNil bool
	}{
		{
			name:    "WhatsApp error",
			err:     waErr,
			wantNil: false,
		},
		{
			name:    "standard error",
			err:     stdErr,
			wantNil: true,
		},
		{
			name:    "nil error",
			err:     nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWhatsAppError(tt.err)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetWhatsAppError() nil = %v, want nil = %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestWhatsAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := NewWhatsAppError(
		ErrCodeSessionNotFound,
		"Session not found",
		http.StatusNotFound,
		underlyingErr,
	)

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlyingErr {
		t.Errorf("Expected unwrapped error to be underlying error")
	}
}

func TestNewWebhookDeliveryFailedError(t *testing.T) {
	url := "https://example.com/webhook"
	attempts := 3
	underlyingErr := errors.New("connection timeout")
	err := NewWebhookDeliveryFailedError(url, attempts, underlyingErr)

	if err.Code != ErrCodeWebhookDeliveryFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeWebhookDeliveryFailed, err.Code)
	}

	if err.HTTPStatus != http.StatusBadGateway {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadGateway, err.HTTPStatus)
	}

	if err.Details["webhook_url"] != url {
		t.Errorf("Expected webhook_url detail to be %s, got %v", url, err.Details["webhook_url"])
	}

	if err.Details["attempts"] != attempts {
		t.Errorf("Expected attempts detail to be %d, got %v", attempts, err.Details["attempts"])
	}
}

func TestNewReconnectionFailedError(t *testing.T) {
	sessionID := "session-123"
	attempts := 5
	underlyingErr := errors.New("connection refused")
	err := NewReconnectionFailedError(sessionID, attempts, underlyingErr)

	if err.Code != ErrCodeReconnectionFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeReconnectionFailed, err.Code)
	}

	if err.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusServiceUnavailable, err.HTTPStatus)
	}

	if err.Details["session_id"] != sessionID {
		t.Errorf("Expected session_id detail to be %s, got %v", sessionID, err.Details["session_id"])
	}

	if err.Details["attempts"] != attempts {
		t.Errorf("Expected attempts detail to be %d, got %v", attempts, err.Details["attempts"])
	}
}
