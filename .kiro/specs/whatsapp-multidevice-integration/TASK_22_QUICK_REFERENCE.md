# Task 22: Error Handling and Logging - Quick Reference

## Error Types Quick Reference

### Creating Errors

```go
import "golang-chi-hexagonal-base/internal/pkg/errors"

// Session errors
err := errors.NewSessionNotFoundError(sessionID)
err := errors.NewSessionNotConnectedError(sessionID)
err := errors.NewSessionAlreadyExistsError(userID, sessionName)
err := errors.NewUnauthorizedSessionError(userID, sessionID)

// Message errors
err := errors.NewMessageSendFailedError(sessionID, recipient, underlyingErr)
err := errors.NewInvalidPhoneNumberError(phoneNumber)

// Media errors
err := errors.NewMediaTooLargeError(size, maxSize)
err := errors.NewInvalidMediaTypeError(mediaType)

// Client errors
err := errors.NewClientNotFoundError(sessionID)
err := errors.NewClientConnectionFailedError(sessionID, underlyingErr)

// QR Code errors
err := errors.NewQRCodeExpiredError(sessionID)
err := errors.NewQRCodeGenerationFailedError(sessionID, underlyingErr)

// Generic errors
err := errors.NewWhatsAppError(code, message, httpStatus, underlyingErr)
```

### Adding Context to Errors

```go
err := errors.NewSessionNotFoundError(sessionID).
    WithDetail("user_id", userID).
    WithDetail("operation", "send_message")

err := errors.NewMessageSendFailedError(sessionID, recipient, underlyingErr).
    WithDetails(map[string]interface{}{
        "message_type": "text",
        "content_length": len(content),
    })
```

### Checking Error Types

```go
// Check if it's a WhatsAppError
if errors.IsWhatsAppError(err) {
    waErr := errors.GetWhatsAppError(err)
    fmt.Printf("Code: %s, Status: %d\n", waErr.Code, waErr.HTTPStatus)
}
```

## Logging Quick Reference

### Initializing Logger

```go
import "golang-chi-hexagonal-base/internal/pkg/logger"

// In service initialization
log := logger.NewFromConfig("info", "json")
whatsappLogger := logger.NewWhatsAppLogger(log)
```

### Session Lifecycle Logging

```go
// Session created
whatsappLogger.LogSessionCreated(sessionID, userID, sessionName)

// Session connected
whatsappLogger.LogSessionConnected(sessionID, phoneNumber)

// Session disconnected
whatsappLogger.LogSessionDisconnected(sessionID, reason)

// Session deleted
whatsappLogger.LogSessionDeleted(sessionID, userID)

// Status update
whatsappLogger.LogSessionStatusUpdate(sessionID, oldStatus, newStatus)
```

### Message Logging

```go
// Message sending
whatsappLogger.LogMessageSending(sessionID, messageID, recipient, messageType)

// Message sent successfully
whatsappLogger.LogMessageSent(sessionID, messageID, recipient, duration)

// Message send failed
whatsappLogger.LogMessageSendFailed(sessionID, messageID, recipient, err)

// Message received
whatsappLogger.LogMessageReceived(sessionID, messageID, sender, messageType)

// Message stored
whatsappLogger.LogMessageStored(sessionID, messageID, direction)
```

### QR Code Logging

```go
// QR code generated
whatsappLogger.LogQRCodeGenerated(sessionID, expiryTime)

// QR code expired
whatsappLogger.LogQRCodeExpired(sessionID)

// QR code scanned
whatsappLogger.LogQRCodeScanned(sessionID)
```

### Media Logging

```go
// Upload started
whatsappLogger.LogMediaUploadStarted(sessionID, messageID, mediaType, size)

// Upload completed
whatsappLogger.LogMediaUploadCompleted(sessionID, messageID, mediaType, duration)

// Upload failed
whatsappLogger.LogMediaUploadFailed(sessionID, messageID, mediaType, err)
```

### Webhook Logging

```go
// Delivery started
whatsappLogger.LogWebhookDeliveryStarted(sessionID, messageID, webhookURL, attempt)

// Delivery success
whatsappLogger.LogWebhookDeliverySuccess(sessionID, messageID, webhookURL, statusCode, duration)

// Delivery failed
whatsappLogger.LogWebhookDeliveryFailed(sessionID, messageID, webhookURL, attempt, statusCode, err)

// Max retries reached
whatsappLogger.LogWebhookMaxRetriesReached(sessionID, messageID, webhookURL, attempts)
```

### Reconnection Logging

```go
// Reconnection started
whatsappLogger.LogReconnectionStarted(sessionID, attempt, maxAttempts)

// Reconnection success
whatsappLogger.LogReconnectionSuccess(sessionID, attempt, duration)

// Reconnection failed
whatsappLogger.LogReconnectionFailed(sessionID, attempt, maxAttempts, err)

// Max attempts reached
whatsappLogger.LogReconnectionMaxAttemptsReached(sessionID, attempts)
```

### Generic Logging

```go
// Error with context
whatsappLogger.LogErrorWithContext("operation_name", err, map[string]interface{}{
    "session_id": sessionID,
    "user_id": userID,
    "additional_info": "value",
})

// Warning with context
whatsappLogger.LogWarningWithContext("operation_name", "warning message", map[string]interface{}{
    "session_id": sessionID,
    "reason": "some reason",
})

// Performance metric
whatsappLogger.LogPerformanceMetric("operation_name", duration, map[string]interface{}{
    "session_id": sessionID,
    "items_processed": count,
})
```

### Context Helpers

```go
// Add session context
sessionLogger := whatsappLogger.WithSessionContext(sessionID, userID)
sessionLogger.LogMessageSent(sessionID, messageID, recipient, duration)

// Add message context
messageLogger := whatsappLogger.WithMessageContext(sessionID, messageID)
messageLogger.LogWebhookDeliverySuccess(sessionID, messageID, webhookURL, statusCode, duration)
```

## Common Patterns

### Service Method Pattern

```go
func (s *WhatsAppService) SomeOperation(ctx context.Context, sessionID, userID string) error {
    // Log operation start
    s.logger.LogWarningWithContext("some_operation", "starting operation", map[string]interface{}{
        "session_id": sessionID,
        "user_id": userID,
    })
    
    // Track performance
    startTime := time.Now()
    defer func() {
        s.logger.LogPerformanceMetric("some_operation", time.Since(startTime), map[string]interface{}{
            "session_id": sessionID,
        })
    }()
    
    // Perform operation
    result, err := s.doSomething(ctx, sessionID)
    if err != nil {
        s.logger.LogErrorWithContext("some_operation", err, map[string]interface{}{
            "session_id": sessionID,
            "user_id": userID,
        })
        return errors.NewSomeSpecificError(sessionID, err)
    }
    
    // Log success
    s.logger.LogWarningWithContext("some_operation", "operation completed", map[string]interface{}{
        "session_id": sessionID,
        "result_count": len(result),
    })
    
    return nil
}
```

### Handler Error Handling Pattern

```go
func (h *WhatsAppHandler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    // ... extract parameters ...
    
    // Call service
    result, err := h.whatsappService.SomeOperation(r.Context(), sessionID, userID)
    if err != nil {
        handleWhatsAppError(w, err)  // Centralized error handling
        return
    }
    
    // Return success response
    respondSuccess(w, http.StatusOK, "Operation successful", result)
}
```

## Error Code Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `WHATSAPP_SESSION_NOT_FOUND` | 404 | Session doesn't exist |
| `WHATSAPP_SESSION_NOT_CONNECTED` | 400 | Session not connected |
| `WHATSAPP_SESSION_ALREADY_EXISTS` | 409 | Duplicate session name |
| `WHATSAPP_UNAUTHORIZED_SESSION` | 403 | No access to session |
| `WHATSAPP_QR_CODE_EXPIRED` | 410 | QR code expired |
| `WHATSAPP_MESSAGE_SEND_FAILED` | 500 | Message send failed |
| `WHATSAPP_INVALID_PHONE_NUMBER` | 400 | Invalid phone format |
| `WHATSAPP_MEDIA_TOO_LARGE` | 413 | File too large |
| `WHATSAPP_INVALID_MEDIA_TYPE` | 400 | Unsupported media type |
| `WHATSAPP_CLIENT_NOT_FOUND` | 404 | Client doesn't exist |
| `WHATSAPP_SERVICE_UNAVAILABLE` | 503 | Service down |
| `WHATSAPP_RATE_LIMIT_EXCEEDED` | 429 | Too many requests |

## Log Event Types

| Event | Level | Description |
|-------|-------|-------------|
| `session_created` | INFO | New session created |
| `session_connected` | INFO | Session connected to WhatsApp |
| `session_disconnected` | WARN | Session disconnected |
| `message_sent` | INFO | Message sent successfully |
| `message_send_failed` | ERROR | Message send failed |
| `message_received` | INFO | Incoming message |
| `qr_code_generated` | INFO | QR code created |
| `webhook_delivery_success` | INFO | Webhook delivered |
| `webhook_delivery_failed` | WARN | Webhook failed |
| `reconnection_started` | INFO | Reconnection attempt |
| `reconnection_failed` | WARN | Reconnection failed |
| `error` | ERROR | Generic error |
| `performance_metric` | DEBUG | Performance data |

## Troubleshooting

### Finding Errors in Logs

```bash
# All WhatsApp errors
jq 'select(.component == "whatsapp" and .level == "ERROR")' logs/app.log

# Specific error code
jq 'select(.component == "whatsapp" and .error_code == "WHATSAPP_SESSION_NOT_FOUND")' logs/app.log

# Errors for specific session
jq 'select(.component == "whatsapp" and .session_id == "uuid-123" and .level == "ERROR")' logs/app.log
```

### Performance Analysis

```bash
# Slow operations (>1s)
jq 'select(.event == "performance_metric" and .duration_ms > 1000)' logs/app.log

# Average message send time
jq 'select(.event == "message_sent") | .duration_ms' logs/app.log | jq -s 'add/length'
```

### Webhook Monitoring

```bash
# Failed webhooks
jq 'select(.event == "webhook_delivery_failed")' logs/app.log

# Webhook retry patterns
jq 'select(.event == "webhook_retrying")' logs/app.log
```

## Best Practices

1. **Always log operation start and completion**
2. **Include relevant IDs in all log entries** (session_id, user_id, message_id)
3. **Use appropriate log levels** (DEBUG for details, INFO for normal, WARN for issues, ERROR for failures)
4. **Track performance for critical operations**
5. **Log errors with full context** (what, when, who, why)
6. **Use structured errors** with proper codes and HTTP status
7. **Chain errors** to preserve underlying causes
8. **Add context to errors** before returning them
9. **Centralize error handling** in handlers
10. **Monitor error rates** and set up alerts

## Files Reference

- **Error Types**: `internal/pkg/errors/whatsapp_errors.go`
- **Logger**: `internal/pkg/logger/whatsapp_logger.go`
- **Service**: `internal/domain/service/whatsapp_service.go`
- **Handler**: `internal/adapter/http/handler/whatsapp_handler.go`
