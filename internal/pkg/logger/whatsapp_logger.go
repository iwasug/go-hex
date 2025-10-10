package logger

import (
	"fmt"
	"time"
)

// WhatsAppLogger provides structured logging for WhatsApp operations
type WhatsAppLogger struct {
	logger LoggerInterface
}

// NewWhatsAppLogger creates a new WhatsAppLogger
func NewWhatsAppLogger(logger LoggerInterface) *WhatsAppLogger {
	return &WhatsAppLogger{
		logger: logger.WithFields(Fields{
			"component": "whatsapp",
		}),
	}
}

// Session Lifecycle Logging

// LogSessionCreated logs when a session is created
func (l *WhatsAppLogger) LogSessionCreated(sessionID, userID, sessionName string) {
	l.logger.Info("WhatsApp session created", Fields{
		"event":        "session_created",
		"session_id":   sessionID,
		"user_id":      userID,
		"session_name": sessionName,
	})
}

// LogSessionDeleted logs when a session is deleted
func (l *WhatsAppLogger) LogSessionDeleted(sessionID, userID string) {
	l.logger.Info("WhatsApp session deleted", Fields{
		"event":      "session_deleted",
		"session_id": sessionID,
		"user_id":    userID,
	})
}

// LogSessionConnecting logs when a session starts connecting
func (l *WhatsAppLogger) LogSessionConnecting(sessionID string) {
	l.logger.Info("WhatsApp session connecting", Fields{
		"event":      "session_connecting",
		"session_id": sessionID,
	})
}

// LogSessionConnected logs when a session successfully connects
func (l *WhatsAppLogger) LogSessionConnected(sessionID, phoneNumber string) {
	l.logger.Info("WhatsApp session connected", Fields{
		"event":        "session_connected",
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})
}

// LogSessionDisconnected logs when a session disconnects
func (l *WhatsAppLogger) LogSessionDisconnected(sessionID, reason string) {
	l.logger.Warn("WhatsApp session disconnected", Fields{
		"event":      "session_disconnected",
		"session_id": sessionID,
		"reason":     reason,
	})
}

// LogSessionAuthenticating logs when a session is authenticating
func (l *WhatsAppLogger) LogSessionAuthenticating(sessionID string) {
	l.logger.Info("WhatsApp session authenticating", Fields{
		"event":      "session_authenticating",
		"session_id": sessionID,
	})
}

// LogSessionAuthSuccess logs successful authentication
func (l *WhatsAppLogger) LogSessionAuthSuccess(sessionID, phoneNumber string) {
	l.logger.Info("WhatsApp session authenticated successfully", Fields{
		"event":        "session_auth_success",
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})
}

// LogSessionAuthFailed logs failed authentication
func (l *WhatsAppLogger) LogSessionAuthFailed(sessionID string, err error) {
	l.logger.Error("WhatsApp session authentication failed", Fields{
		"event":      "session_auth_failed",
		"session_id": sessionID,
		"error":      err.Error(),
	})
}

// LogSessionStatusUpdate logs when a session status changes
func (l *WhatsAppLogger) LogSessionStatusUpdate(sessionID, oldStatus, newStatus string) {
	l.logger.Info("WhatsApp session status updated", Fields{
		"event":      "session_status_update",
		"session_id": sessionID,
		"old_status": oldStatus,
		"new_status": newStatus,
	})
}

// QR Code Logging

// LogQRCodeGenerated logs when a QR code is generated
func (l *WhatsAppLogger) LogQRCodeGenerated(sessionID string, expiryTime time.Time) {
	l.logger.Info("WhatsApp QR code generated", Fields{
		"event":       "qr_code_generated",
		"session_id":  sessionID,
		"expiry_time": expiryTime.Format(time.RFC3339),
	})
}

// LogQRCodeExpired logs when a QR code expires
func (l *WhatsAppLogger) LogQRCodeExpired(sessionID string) {
	l.logger.Warn("WhatsApp QR code expired", Fields{
		"event":      "qr_code_expired",
		"session_id": sessionID,
	})
}

// LogQRCodeScanned logs when a QR code is scanned
func (l *WhatsAppLogger) LogQRCodeScanned(sessionID string) {
	l.logger.Info("WhatsApp QR code scanned", Fields{
		"event":      "qr_code_scanned",
		"session_id": sessionID,
	})
}

// Message Logging

// LogMessageSending logs when a message is being sent
func (l *WhatsAppLogger) LogMessageSending(sessionID, messageID, recipient, messageType string) {
	l.logger.Info("Sending WhatsApp message", Fields{
		"event":        "message_sending",
		"session_id":   sessionID,
		"message_id":   messageID,
		"recipient":    recipient,
		"message_type": messageType,
	})
}

// LogMessageSent logs when a message is successfully sent
func (l *WhatsAppLogger) LogMessageSent(sessionID, messageID, recipient string, duration time.Duration) {
	l.logger.Info("WhatsApp message sent successfully", Fields{
		"event":      "message_sent",
		"session_id": sessionID,
		"message_id": messageID,
		"recipient":  recipient,
		"duration_ms": duration.Milliseconds(),
	})
}

// LogMessageSendFailed logs when a message fails to send
func (l *WhatsAppLogger) LogMessageSendFailed(sessionID, messageID, recipient string, err error) {
	l.logger.Error("Failed to send WhatsApp message", Fields{
		"event":      "message_send_failed",
		"session_id": sessionID,
		"message_id": messageID,
		"recipient":  recipient,
		"error":      err.Error(),
	})
}

// LogMessageReceived logs when a message is received
func (l *WhatsAppLogger) LogMessageReceived(sessionID, messageID, sender, messageType string) {
	l.logger.Info("WhatsApp message received", Fields{
		"event":        "message_received",
		"session_id":   sessionID,
		"message_id":   messageID,
		"sender":       sender,
		"message_type": messageType,
	})
}

// LogMessageStored logs when a message is stored in the database
func (l *WhatsAppLogger) LogMessageStored(sessionID, messageID, direction string) {
	l.logger.Debug("WhatsApp message stored", Fields{
		"event":      "message_stored",
		"session_id": sessionID,
		"message_id": messageID,
		"direction":  direction,
	})
}

// LogMessageStorageFailed logs when message storage fails
func (l *WhatsAppLogger) LogMessageStorageFailed(sessionID, messageID string, err error) {
	l.logger.Error("Failed to store WhatsApp message", Fields{
		"event":      "message_storage_failed",
		"session_id": sessionID,
		"message_id": messageID,
		"error":      err.Error(),
	})
}

// LogMessageStatusUpdate logs when a message status changes
func (l *WhatsAppLogger) LogMessageStatusUpdate(sessionID, messageID, oldStatus, newStatus string) {
	l.logger.Debug("WhatsApp message status updated", Fields{
		"event":      "message_status_update",
		"session_id": sessionID,
		"message_id": messageID,
		"old_status": oldStatus,
		"new_status": newStatus,
	})
}

// Media Logging

// LogMediaUploadStarted logs when media upload starts
func (l *WhatsAppLogger) LogMediaUploadStarted(sessionID, messageID, mediaType string, size int64) {
	l.logger.Info("WhatsApp media upload started", Fields{
		"event":       "media_upload_started",
		"session_id":  sessionID,
		"message_id":  messageID,
		"media_type":  mediaType,
		"size_bytes":  size,
	})
}

// LogMediaUploadCompleted logs when media upload completes
func (l *WhatsAppLogger) LogMediaUploadCompleted(sessionID, messageID, mediaType string, duration time.Duration) {
	l.logger.Info("WhatsApp media upload completed", Fields{
		"event":       "media_upload_completed",
		"session_id":  sessionID,
		"message_id":  messageID,
		"media_type":  mediaType,
		"duration_ms": duration.Milliseconds(),
	})
}

// LogMediaUploadFailed logs when media upload fails
func (l *WhatsAppLogger) LogMediaUploadFailed(sessionID, messageID, mediaType string, err error) {
	l.logger.Error("WhatsApp media upload failed", Fields{
		"event":      "media_upload_failed",
		"session_id": sessionID,
		"message_id": messageID,
		"media_type": mediaType,
		"error":      err.Error(),
	})
}

// Webhook Logging

// LogWebhookDeliveryStarted logs when webhook delivery starts
func (l *WhatsAppLogger) LogWebhookDeliveryStarted(sessionID, messageID, webhookURL string, attempt int) {
	l.logger.Info("WhatsApp webhook delivery started", Fields{
		"event":       "webhook_delivery_started",
		"session_id":  sessionID,
		"message_id":  messageID,
		"webhook_url": webhookURL,
		"attempt":     attempt,
	})
}

// LogWebhookDeliverySuccess logs successful webhook delivery
func (l *WhatsAppLogger) LogWebhookDeliverySuccess(sessionID, messageID, webhookURL string, statusCode int, duration time.Duration) {
	l.logger.Info("WhatsApp webhook delivered successfully", Fields{
		"event":       "webhook_delivery_success",
		"session_id":  sessionID,
		"message_id":  messageID,
		"webhook_url": webhookURL,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	})
}

// LogWebhookDeliveryFailed logs failed webhook delivery
func (l *WhatsAppLogger) LogWebhookDeliveryFailed(sessionID, messageID, webhookURL string, attempt int, statusCode int, err error) {
	fields := Fields{
		"event":       "webhook_delivery_failed",
		"session_id":  sessionID,
		"message_id":  messageID,
		"webhook_url": webhookURL,
		"attempt":     attempt,
	}
	
	if statusCode > 0 {
		fields["status_code"] = statusCode
	}
	
	if err != nil {
		fields["error"] = err.Error()
	}
	
	l.logger.Warn("WhatsApp webhook delivery failed", fields)
}

// LogWebhookRetrying logs when webhook delivery is being retried
func (l *WhatsAppLogger) LogWebhookRetrying(sessionID, messageID, webhookURL string, attempt, maxAttempts int, delay time.Duration) {
	l.logger.Info("Retrying WhatsApp webhook delivery", Fields{
		"event":        "webhook_retrying",
		"session_id":   sessionID,
		"message_id":   messageID,
		"webhook_url":  webhookURL,
		"attempt":      attempt,
		"max_attempts": maxAttempts,
		"delay_ms":     delay.Milliseconds(),
	})
}

// LogWebhookMaxRetriesReached logs when max webhook retries are reached
func (l *WhatsAppLogger) LogWebhookMaxRetriesReached(sessionID, messageID, webhookURL string, attempts int) {
	l.logger.Error("WhatsApp webhook max retries reached", Fields{
		"event":       "webhook_max_retries_reached",
		"session_id":  sessionID,
		"message_id":  messageID,
		"webhook_url": webhookURL,
		"attempts":    attempts,
	})
}

// Reconnection Logging

// LogReconnectionStarted logs when reconnection starts
func (l *WhatsAppLogger) LogReconnectionStarted(sessionID string, attempt, maxAttempts int) {
	l.logger.Info("WhatsApp session reconnection started", Fields{
		"event":        "reconnection_started",
		"session_id":   sessionID,
		"attempt":      attempt,
		"max_attempts": maxAttempts,
	})
}

// LogReconnectionSuccess logs successful reconnection
func (l *WhatsAppLogger) LogReconnectionSuccess(sessionID string, attempt int, duration time.Duration) {
	l.logger.Info("WhatsApp session reconnected successfully", Fields{
		"event":       "reconnection_success",
		"session_id":  sessionID,
		"attempt":     attempt,
		"duration_ms": duration.Milliseconds(),
	})
}

// LogReconnectionFailed logs failed reconnection attempt
func (l *WhatsAppLogger) LogReconnectionFailed(sessionID string, attempt, maxAttempts int, err error) {
	l.logger.Warn("WhatsApp session reconnection attempt failed", Fields{
		"event":        "reconnection_failed",
		"session_id":   sessionID,
		"attempt":      attempt,
		"max_attempts": maxAttempts,
		"error":        err.Error(),
	})
}

// LogReconnectionMaxAttemptsReached logs when max reconnection attempts are reached
func (l *WhatsAppLogger) LogReconnectionMaxAttemptsReached(sessionID string, attempts int) {
	l.logger.Error("WhatsApp session max reconnection attempts reached", Fields{
		"event":      "reconnection_max_attempts_reached",
		"session_id": sessionID,
		"attempts":   attempts,
	})
}

// Client Management Logging

// LogClientCreated logs when a WhatsApp client is created
func (l *WhatsAppLogger) LogClientCreated(sessionID, userID string) {
	l.logger.Info("WhatsApp client created", Fields{
		"event":      "client_created",
		"session_id": sessionID,
		"user_id":    userID,
	})
}

// LogClientDeleted logs when a WhatsApp client is deleted
func (l *WhatsAppLogger) LogClientDeleted(sessionID string) {
	l.logger.Info("WhatsApp client deleted", Fields{
		"event":      "client_deleted",
		"session_id": sessionID,
	})
}

// LogClientRestoreStarted logs when client restoration starts
func (l *WhatsAppLogger) LogClientRestoreStarted(sessionCount int) {
	l.logger.Info("WhatsApp client restoration started", Fields{
		"event":         "client_restore_started",
		"session_count": sessionCount,
	})
}

// LogClientRestoreCompleted logs when client restoration completes
func (l *WhatsAppLogger) LogClientRestoreCompleted(successCount, failureCount int, duration time.Duration) {
	l.logger.Info("WhatsApp client restoration completed", Fields{
		"event":         "client_restore_completed",
		"success_count": successCount,
		"failure_count": failureCount,
		"duration_ms":   duration.Milliseconds(),
	})
}

// LogClientRestoreFailed logs when a client fails to restore
func (l *WhatsAppLogger) LogClientRestoreFailed(sessionID string, err error) {
	l.logger.Error("WhatsApp client restoration failed", Fields{
		"event":      "client_restore_failed",
		"session_id": sessionID,
		"error":      err.Error(),
	})
}

// Store Operations Logging

// LogStoreOperation logs a store operation
func (l *WhatsAppLogger) LogStoreOperation(operation, sessionID string, success bool, duration time.Duration) {
	level := "info"
	if !success {
		level = "error"
	}
	
	fields := Fields{
		"event":       "store_operation",
		"operation":   operation,
		"session_id":  sessionID,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("WhatsApp store operation: %s", operation)
	
	if level == "error" {
		l.logger.Error(message, fields)
	} else {
		l.logger.Debug(message, fields)
	}
}

// LogStoreError logs a store error
func (l *WhatsAppLogger) LogStoreError(operation, sessionID string, err error) {
	l.logger.Error("WhatsApp store operation failed", Fields{
		"event":      "store_error",
		"operation":  operation,
		"session_id": sessionID,
		"error":      err.Error(),
	})
}

// Error Context Logging

// LogErrorWithContext logs an error with full context
func (l *WhatsAppLogger) LogErrorWithContext(operation string, err error, context map[string]interface{}) {
	fields := Fields{
		"event":     "error",
		"operation": operation,
		"error":     err.Error(),
	}
	
	// Merge context into fields
	for k, v := range context {
		fields[k] = v
	}
	
	l.logger.Error(fmt.Sprintf("WhatsApp operation failed: %s", operation), fields)
}

// LogWarningWithContext logs a warning with context
func (l *WhatsAppLogger) LogWarningWithContext(operation string, message string, context map[string]interface{}) {
	fields := Fields{
		"event":     "warning",
		"operation": operation,
	}
	
	// Merge context into fields
	for k, v := range context {
		fields[k] = v
	}
	
	l.logger.Warn(fmt.Sprintf("WhatsApp warning: %s - %s", operation, message), fields)
}

// Performance Logging

// LogPerformanceMetric logs a performance metric
func (l *WhatsAppLogger) LogPerformanceMetric(operation string, duration time.Duration, context map[string]interface{}) {
	fields := Fields{
		"event":       "performance_metric",
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
	}
	
	// Merge context into fields
	for k, v := range context {
		fields[k] = v
	}
	
	l.logger.Debug(fmt.Sprintf("WhatsApp performance: %s", operation), fields)
}

// WithSessionContext returns a logger with session context
func (l *WhatsAppLogger) WithSessionContext(sessionID, userID string) *WhatsAppLogger {
	return &WhatsAppLogger{
		logger: l.logger.WithFields(Fields{
			"session_id": sessionID,
			"user_id":    userID,
		}),
	}
}

// WithMessageContext returns a logger with message context
func (l *WhatsAppLogger) WithMessageContext(sessionID, messageID string) *WhatsAppLogger {
	return &WhatsAppLogger{
		logger: l.logger.WithFields(Fields{
			"session_id": sessionID,
			"message_id": messageID,
		}),
	}
}
