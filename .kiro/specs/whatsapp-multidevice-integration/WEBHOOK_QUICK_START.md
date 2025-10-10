# Webhook Integration - Quick Start Guide

## Overview

This guide provides a quick reference for implementing and using the webhook integration for WhatsApp messages.

## Setup

### 1. Configure Environment Variables

Add to your `.env` file:

```env
WEBHOOK_TIMEOUT=10s
WEBHOOK_MAX_RETRIES=3
WEBHOOK_RETRY_DELAY=5s
WEBHOOK_MAX_RETRY_DELAY=60s
```

### 2. Initialize Webhook Client in Application

```go
import (
    "golang-chi-hexagonal-base/internal/adapter/webhook"
    "golang-chi-hexagonal-base/internal/pkg/webhook"
)

// In your main.go or dependency injection setup
webhookConfig := webhook.Config{
    Timeout:       config.Webhook.Timeout,
    MaxRetries:    config.Webhook.MaxRetries,
    RetryDelay:    config.Webhook.RetryDelay,
    MaxRetryDelay: config.Webhook.MaxRetryDelay,
}

webhookClient := webhook.NewClientAdapter(webhookConfig, logger)

// Pass to WhatsApp service
whatsappService := service.NewWhatsAppService(
    sessionRepo,
    messageRepo,
    userRepo,
    clientManager,
    webhookClient, // Add webhook client here
)
```

## Usage

### Creating a Session with Webhook

```go
input := service.CreateSessionInput{
    SessionName:   "My WhatsApp Session",
    WebhookURL:    "https://your-domain.com/webhook",
    WebhookSecret: "your-secret-key", // Optional but recommended
}

session, err := whatsappService.CreateSession(ctx, userID, input)
if err != nil {
    // Handle error
}
```

### Webhook Payload Format

When a message is received, your webhook endpoint will receive:

```json
{
  "event": "message.received",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-10-09T12:00:00Z",
  "data": {
    "message_id": "3EB0C767D71D8D5A2A6B",
    "from": "+1234567890",
    "to": "+0987654321",
    "content": "Hello, World!",
    "message_type": "text",
    "timestamp": "2025-10-09T12:00:00Z",
    "direction": "inbound"
  }
}
```

### Implementing a Webhook Receiver

```go
package main

import (
    "encoding/json"
    "io"
    "net/http"
    
    "golang-chi-hexagonal-base/internal/pkg/webhook"
)

type WebhookPayload struct {
    Event     string                 `json:"event"`
    SessionID string                 `json:"session_id"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    // 1. Read the request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // 2. Verify HMAC signature (if using webhook secret)
    signature := r.Header.Get("X-Webhook-Signature")
    secret := "your-secret-key" // Same as configured in session
    
    if !webhook.VerifyWebhookSignature(body, signature, secret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // 3. Parse the payload
    var payload WebhookPayload
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // 4. Process the webhook
    switch payload.Event {
    case "message.received":
        handleMessageReceived(payload)
    default:
        // Unknown event type
    }
    
    // 5. Return success response quickly
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}

func handleMessageReceived(payload WebhookPayload) {
    // Process message asynchronously
    go func() {
        messageID := payload.Data["message_id"].(string)
        from := payload.Data["from"].(string)
        content := payload.Data["content"].(string)
        
        // Your business logic here
        // - Store in database
        // - Trigger notifications
        // - Process commands
        // - etc.
    }()
}
```

## Webhook Headers

Your webhook endpoint will receive these headers:

```
Content-Type: application/json
User-Agent: WhatsApp-Webhook-Client/1.0
X-Webhook-Event: message.received
X-Webhook-Timestamp: 2025-10-09T12:00:00Z
X-Webhook-Signature: <hmac-sha256-hex> (if secret configured)
```

## Best Practices

### 1. Return 200 OK Quickly

```go
// ✅ Good: Return immediately, process async
w.WriteHeader(http.StatusOK)
go processWebhook(payload)

// ❌ Bad: Process before responding
processWebhook(payload) // This might timeout
w.WriteHeader(http.StatusOK)
```

### 2. Verify HMAC Signature

```go
// Always verify signature if using secrets
if !webhook.VerifyWebhookSignature(body, signature, secret) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

### 3. Handle Idempotency

```go
// Check if message already processed
if alreadyProcessed(messageID) {
    w.WriteHeader(http.StatusOK) // Still return success
    return
}

// Process message
processMessage(payload)
```

### 4. Use HTTPS in Production

```go
// ✅ Production
WebhookURL: "https://your-domain.com/webhook"

// ❌ Development only
WebhookURL: "http://localhost:8080/webhook"
```

### 5. Monitor Webhook Failures

```sql
-- Check failed webhook deliveries
SELECT * FROM whatsapp_messages 
WHERE webhook_delivered = false 
AND webhook_attempts >= 3
ORDER BY created_at DESC;
```

## Testing

### Test Webhook Locally

Use a tool like ngrok to expose your local server:

```bash
# Start ngrok
ngrok http 8080

# Use the ngrok URL as webhook URL
https://abc123.ngrok.io/webhook
```

### Test Webhook Signature

```go
package main

import (
    "testing"
    "golang-chi-hexagonal-base/internal/pkg/webhook"
)

func TestWebhookSignature(t *testing.T) {
    secret := "test-secret"
    payload := []byte(`{"event":"test"}`)
    
    // Generate signature (simulating sender)
    client := webhook.NewClient(webhook.DefaultConfig(), logger)
    signature := client.generateHMACSignature(payload, secret)
    
    // Verify signature (simulating receiver)
    if !webhook.VerifyWebhookSignature(payload, signature, secret) {
        t.Error("Signature verification failed")
    }
}
```

### Test Webhook Endpoint

```bash
# Send test webhook
curl -X POST https://your-domain.com/webhook \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Event: message.received" \
  -H "X-Webhook-Signature: <signature>" \
  -d '{
    "event": "message.received",
    "session_id": "test-session",
    "timestamp": "2025-10-09T12:00:00Z",
    "data": {
      "message_id": "test-123",
      "from": "+1234567890",
      "content": "Test message"
    }
  }'
```

## Troubleshooting

### Webhooks Not Being Delivered

**Check:**
1. Webhook URL is accessible from your server
2. Endpoint returns 2xx status code
3. No firewall blocking outbound requests
4. Check logs for error messages

```bash
# Search logs for webhook errors
grep "Webhook delivery failed" logs.json
```

### Signature Verification Fails

**Check:**
1. Secret matches on both sides
2. Using raw request body (not parsed JSON)
3. Signature is hex-encoded
4. No whitespace in signature

```go
// ✅ Correct: Use raw body
body, _ := io.ReadAll(r.Body)
webhook.VerifyWebhookSignature(body, signature, secret)

// ❌ Wrong: Use parsed JSON
var payload map[string]interface{}
json.Unmarshal(body, &payload)
// Can't verify signature after parsing
```

### Webhook Timeouts

**Solutions:**
1. Increase timeout: `WEBHOOK_TIMEOUT=30s`
2. Optimize webhook endpoint performance
3. Return 200 OK immediately, process async
4. Check network latency

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBHOOK_TIMEOUT` | `10s` | HTTP request timeout |
| `WEBHOOK_MAX_RETRIES` | `3` | Maximum retry attempts |
| `WEBHOOK_RETRY_DELAY` | `5s` | Initial retry delay |
| `WEBHOOK_MAX_RETRY_DELAY` | `60s` | Maximum retry delay |

## Retry Behavior

| Attempt | Delay | Total Time |
|---------|-------|------------|
| 1 | 0s | 0s |
| 2 | 5s | 5s |
| 3 | 10s | 15s |

**Total maximum time:** ~15 seconds (plus request timeouts)

## HTTP Status Codes

| Status | Retry? | Description |
|--------|--------|-------------|
| 200-299 | No | Success |
| 400-499 | No | Client error (except 429) |
| 429 | Yes | Rate limit |
| 500-599 | Yes | Server error |
| Network error | Yes | Connection failed |

## Security Checklist

- [ ] Use HTTPS for webhook URLs in production
- [ ] Configure webhook secret for HMAC authentication
- [ ] Verify HMAC signature on receiver side
- [ ] Store webhook secrets encrypted in database
- [ ] Use constant-time comparison for signatures
- [ ] Implement rate limiting on webhook receiver
- [ ] Monitor for suspicious webhook activity
- [ ] Rotate webhook secrets periodically

## Next Steps

1. **Set up webhook endpoint** on your server
2. **Configure webhook URL** when creating WhatsApp session
3. **Implement signature verification** in your receiver
4. **Test with ngrok** or similar tool
5. **Monitor webhook deliveries** in production
6. **Set up alerts** for failed webhooks

## Additional Resources

- Full documentation: `WEBHOOK_INTEGRATION.md`
- Implementation details: `TASK_12_SUMMARY.md`
- Verification report: `TASK_12_VERIFICATION.md`
- Code: `internal/pkg/webhook/client.go`
