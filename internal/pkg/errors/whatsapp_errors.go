package errors

import (
	"fmt"
	"net/http"
)

// WhatsApp-specific error codes
const (
	// Session errors
	ErrCodeSessionNotFound        = "WHATSAPP_SESSION_NOT_FOUND"
	ErrCodeSessionNotConnected    = "WHATSAPP_SESSION_NOT_CONNECTED"
	ErrCodeSessionAlreadyExists   = "WHATSAPP_SESSION_ALREADY_EXISTS"
	ErrCodeSessionCreationFailed  = "WHATSAPP_SESSION_CREATION_FAILED"
	ErrCodeSessionUpdateFailed    = "WHATSAPP_SESSION_UPDATE_FAILED"
	ErrCodeSessionDeleteFailed    = "WHATSAPP_SESSION_DELETE_FAILED"
	ErrCodeSessionRestoreFailed   = "WHATSAPP_SESSION_RESTORE_FAILED"
	ErrCodeSessionDisconnected    = "WHATSAPP_SESSION_DISCONNECTED"
	ErrCodeSessionAuthFailed      = "WHATSAPP_SESSION_AUTH_FAILED"
	ErrCodeSessionTimeout         = "WHATSAPP_SESSION_TIMEOUT"

	// Authorization errors
	ErrCodeUnauthorizedSession = "WHATSAPP_UNAUTHORIZED_SESSION"
	ErrCodeInsufficientPermissions = "WHATSAPP_INSUFFICIENT_PERMISSIONS"

	// QR Code errors
	ErrCodeQRCodeExpired       = "WHATSAPP_QR_CODE_EXPIRED"
	ErrCodeQRCodeGenerationFailed = "WHATSAPP_QR_CODE_GENERATION_FAILED"

	// Message errors
	ErrCodeMessageSendFailed      = "WHATSAPP_MESSAGE_SEND_FAILED"
	ErrCodeMessageReceiveFailed   = "WHATSAPP_MESSAGE_RECEIVE_FAILED"
	ErrCodeMessageStorageFailed   = "WHATSAPP_MESSAGE_STORAGE_FAILED"
	ErrCodeMessageNotFound        = "WHATSAPP_MESSAGE_NOT_FOUND"
	ErrCodeInvalidMessageType     = "WHATSAPP_INVALID_MESSAGE_TYPE"
	ErrCodeMessageTooLarge        = "WHATSAPP_MESSAGE_TOO_LARGE"

	// Phone number errors
	ErrCodeInvalidPhoneNumber = "WHATSAPP_INVALID_PHONE_NUMBER"
	ErrCodePhoneNotRegistered = "WHATSAPP_PHONE_NOT_REGISTERED"

	// Media errors
	ErrCodeMediaUploadFailed    = "WHATSAPP_MEDIA_UPLOAD_FAILED"
	ErrCodeMediaDownloadFailed  = "WHATSAPP_MEDIA_DOWNLOAD_FAILED"
	ErrCodeMediaTooLarge        = "WHATSAPP_MEDIA_TOO_LARGE"
	ErrCodeInvalidMediaType     = "WHATSAPP_INVALID_MEDIA_TYPE"
	ErrCodeMediaStorageFailed   = "WHATSAPP_MEDIA_STORAGE_FAILED"

	// Webhook errors
	ErrCodeWebhookDeliveryFailed = "WHATSAPP_WEBHOOK_DELIVERY_FAILED"
	ErrCodeWebhookTimeout        = "WHATSAPP_WEBHOOK_TIMEOUT"
	ErrCodeWebhookInvalidURL     = "WHATSAPP_WEBHOOK_INVALID_URL"

	// Client errors
	ErrCodeClientNotFound        = "WHATSAPP_CLIENT_NOT_FOUND"
	ErrCodeClientCreationFailed  = "WHATSAPP_CLIENT_CREATION_FAILED"
	ErrCodeClientConnectionFailed = "WHATSAPP_CLIENT_CONNECTION_FAILED"
	ErrCodeClientDisconnectFailed = "WHATSAPP_CLIENT_DISCONNECT_FAILED"

	// Service errors
	ErrCodeServiceUnavailable = "WHATSAPP_SERVICE_UNAVAILABLE"
	ErrCodeServiceTimeout     = "WHATSAPP_SERVICE_TIMEOUT"
	ErrCodeRateLimitExceeded  = "WHATSAPP_RATE_LIMIT_EXCEEDED"

	// Store errors
	ErrCodeStoreOperationFailed = "WHATSAPP_STORE_OPERATION_FAILED"
	ErrCodeDeviceNotFound       = "WHATSAPP_DEVICE_NOT_FOUND"
	ErrCodeKeyStorageFailed     = "WHATSAPP_KEY_STORAGE_FAILED"

	// Reconnection errors
	ErrCodeReconnectionFailed    = "WHATSAPP_RECONNECTION_FAILED"
	ErrCodeMaxReconnectAttempts  = "WHATSAPP_MAX_RECONNECT_ATTEMPTS"
)

// WhatsAppError represents a WhatsApp-specific error with additional context
type WhatsAppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Details    map[string]interface{}
	Underlying error
}

// Error implements the error interface
func (e *WhatsAppError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Underlying)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *WhatsAppError) Unwrap() error {
	return e.Underlying
}

// WithDetails adds additional context to the error
func (e *WhatsAppError) WithDetails(details map[string]interface{}) *WhatsAppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithDetail adds a single detail to the error
func (e *WhatsAppError) WithDetail(key string, value interface{}) *WhatsAppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewWhatsAppError creates a new WhatsAppError
func NewWhatsAppError(code, message string, httpStatus int, underlying error) *WhatsAppError {
	return &WhatsAppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    make(map[string]interface{}),
		Underlying: underlying,
	}
}

// Session Errors

// NewSessionNotFoundError creates a session not found error
func NewSessionNotFoundError(sessionID string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeSessionNotFound,
		"WhatsApp session not found",
		http.StatusNotFound,
		nil,
	).WithDetail("session_id", sessionID)
}

// NewSessionNotConnectedError creates a session not connected error
func NewSessionNotConnectedError(sessionID string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeSessionNotConnected,
		"WhatsApp session is not connected",
		http.StatusBadRequest,
		nil,
	).WithDetail("session_id", sessionID)
}

// NewSessionAlreadyExistsError creates a session already exists error
func NewSessionAlreadyExistsError(userID, sessionName string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeSessionAlreadyExists,
		"Session with this name already exists",
		http.StatusConflict,
		nil,
	).WithDetails(map[string]interface{}{
		"user_id":      userID,
		"session_name": sessionName,
	})
}

// NewSessionCreationFailedError creates a session creation failed error
func NewSessionCreationFailedError(err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeSessionCreationFailed,
		"Failed to create WhatsApp session",
		http.StatusInternalServerError,
		err,
	)
}

// NewSessionDisconnectedError creates a session disconnected error
func NewSessionDisconnectedError(sessionID string, reason string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeSessionDisconnected,
		"WhatsApp session disconnected",
		http.StatusServiceUnavailable,
		nil,
	).WithDetails(map[string]interface{}{
		"session_id": sessionID,
		"reason":     reason,
	})
}

// Authorization Errors

// NewUnauthorizedSessionError creates an unauthorized session error
func NewUnauthorizedSessionError(userID, sessionID string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeUnauthorizedSession,
		"You don't have access to this session",
		http.StatusForbidden,
		nil,
	).WithDetails(map[string]interface{}{
		"user_id":    userID,
		"session_id": sessionID,
	})
}

// QR Code Errors

// NewQRCodeExpiredError creates a QR code expired error
func NewQRCodeExpiredError(sessionID string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeQRCodeExpired,
		"QR code has expired, request a new one",
		http.StatusGone,
		nil,
	).WithDetail("session_id", sessionID)
}

// NewQRCodeGenerationFailedError creates a QR code generation failed error
func NewQRCodeGenerationFailedError(sessionID string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeQRCodeGenerationFailed,
		"Failed to generate QR code",
		http.StatusInternalServerError,
		err,
	).WithDetail("session_id", sessionID)
}

// Message Errors

// NewMessageSendFailedError creates a message send failed error
func NewMessageSendFailedError(sessionID, recipient string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeMessageSendFailed,
		"Failed to send message",
		http.StatusInternalServerError,
		err,
	).WithDetails(map[string]interface{}{
		"session_id": sessionID,
		"recipient":  recipient,
	})
}

// NewMessageStorageFailedError creates a message storage failed error
func NewMessageStorageFailedError(messageID string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeMessageStorageFailed,
		"Failed to store message",
		http.StatusInternalServerError,
		err,
	).WithDetail("message_id", messageID)
}

// Phone Number Errors

// NewInvalidPhoneNumberError creates an invalid phone number error
func NewInvalidPhoneNumberError(phoneNumber string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeInvalidPhoneNumber,
		"Invalid phone number format (E.164 required)",
		http.StatusBadRequest,
		nil,
	).WithDetail("phone_number", phoneNumber)
}

// Media Errors

// NewMediaTooLargeError creates a media too large error
func NewMediaTooLargeError(size, maxSize int64) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeMediaTooLarge,
		fmt.Sprintf("Media file too large (max: %d MB)", maxSize/(1024*1024)),
		http.StatusRequestEntityTooLarge,
		nil,
	).WithDetails(map[string]interface{}{
		"size":     size,
		"max_size": maxSize,
	})
}

// NewInvalidMediaTypeError creates an invalid media type error
func NewInvalidMediaTypeError(mediaType string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeInvalidMediaType,
		"Invalid or unsupported media type",
		http.StatusBadRequest,
		nil,
	).WithDetail("media_type", mediaType)
}

// Webhook Errors

// NewWebhookDeliveryFailedError creates a webhook delivery failed error
func NewWebhookDeliveryFailedError(url string, attempts int, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeWebhookDeliveryFailed,
		"Failed to deliver webhook after retries",
		http.StatusBadGateway,
		err,
	).WithDetails(map[string]interface{}{
		"webhook_url": url,
		"attempts":    attempts,
	})
}

// Client Errors

// NewClientNotFoundError creates a client not found error
func NewClientNotFoundError(sessionID string) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeClientNotFound,
		"WhatsApp client not found",
		http.StatusNotFound,
		nil,
	).WithDetail("session_id", sessionID)
}

// NewClientConnectionFailedError creates a client connection failed error
func NewClientConnectionFailedError(sessionID string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeClientConnectionFailed,
		"Failed to connect WhatsApp client",
		http.StatusServiceUnavailable,
		err,
	).WithDetail("session_id", sessionID)
}

// NewClientCreationFailedError creates a client creation failed error
func NewClientCreationFailedError(sessionID string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeClientCreationFailed,
		"Failed to create WhatsApp client",
		http.StatusInternalServerError,
		err,
	).WithDetail("session_id", sessionID)
}

// Service Errors

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeServiceUnavailable,
		"WhatsApp service is unavailable",
		http.StatusServiceUnavailable,
		err,
	)
}

// NewRateLimitExceededError creates a rate limit exceeded error
func NewRateLimitExceededError(retryAfter int) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeRateLimitExceeded,
		"Rate limit exceeded",
		http.StatusTooManyRequests,
		nil,
	).WithDetail("retry_after_seconds", retryAfter)
}

// Store Errors

// NewStoreOperationFailedError creates a store operation failed error
func NewStoreOperationFailedError(operation string, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeStoreOperationFailed,
		fmt.Sprintf("Store operation failed: %s", operation),
		http.StatusInternalServerError,
		err,
	).WithDetail("operation", operation)
}

// Reconnection Errors

// NewReconnectionFailedError creates a reconnection failed error
func NewReconnectionFailedError(sessionID string, attempts int, err error) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeReconnectionFailed,
		"Failed to reconnect session",
		http.StatusServiceUnavailable,
		err,
	).WithDetails(map[string]interface{}{
		"session_id": sessionID,
		"attempts":   attempts,
	})
}

// NewMaxReconnectAttemptsError creates a max reconnect attempts error
func NewMaxReconnectAttemptsError(sessionID string, maxAttempts int) *WhatsAppError {
	return NewWhatsAppError(
		ErrCodeMaxReconnectAttempts,
		"Maximum reconnection attempts reached",
		http.StatusServiceUnavailable,
		nil,
	).WithDetails(map[string]interface{}{
		"session_id":   sessionID,
		"max_attempts": maxAttempts,
	})
}

// IsWhatsAppError checks if an error is a WhatsAppError
func IsWhatsAppError(err error) bool {
	_, ok := err.(*WhatsAppError)
	return ok
}

// GetWhatsAppError extracts a WhatsAppError from an error chain
func GetWhatsAppError(err error) *WhatsAppError {
	if waErr, ok := err.(*WhatsAppError); ok {
		return waErr
	}
	return nil
}
