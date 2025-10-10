# Webhook Integration Documentation

## Overview

The webhook integration provides reliable HTTP delivery of WhatsApp events (primarily incoming messages) to external endpoints. It implements retry logic with exponential backoff, HMAC authentication, and comprehensive error handling.

## Architecture

### Components

1. **Webhook Client** (`internal/pkg/webhook/client.go`)
   - Core HTTP client for webhook delivery
   - Implements retry logic with exponential backoff
   - Generates HMAC-SHA256 signatures for authentication
   - Handles timeouts and context cancellation

2. **Webhook Adapter** (`internal/adapter/webhook/client_adapter.go`)
   - Adapts the webhook client to the domain service interface
   - Provides clean separation between infrastructure and domain layers

3. **WhatsApp Service Integration** (`internal/domain/service/whatsapp_service.go`)
   - Integrates webhook delivery into message handling flow
   - Delivers webhooks asynchronously (non-blocking)
   - Tracks delivery attempts in the database

## Features

### 1. Retry Logic with Exponential Backoff

The webhook client automatically retries failed deliveries with exponential backoff:

- **Initial Delay**: 5 seconds (configurable)
- **Backoff Formula**: `delay * 2^(attempt-1)`
- **Max Delay**: 60 seconds (configurable)
- **Max Retries**: 3 attempts (configurable)

**Example Retry Schedule:**
- Attempt 1: Immediate
- Attempt 2: After 5 seconds
- Attempt 3: After 10 seconds (5 * 2^1)

### 2. HMAC Signature Authentication

Webhooks can be authenticated using HMAC-SHA256 signatures:

- **Algorithm**: HMAC-SHA256
- **Header**: `X-Webhook-Signature`
- **Format**: Hex-encoded signature
- **Payload**: Raw JSON body

**Verification Example:**
```go
import "golang-chi-hexagonal-base/internal/pkg/webhook"

func verifyWebhook(body []byte, signature, secret string) bool {
    return webhook.VerifyWebhookSignature(body, signature, secret)
}
```

### 3. Webhook Delivery Tracking

Message entities track webhook delivery status:

- `webhook_delivered`: Boolean flag indicating successful delivery
- `webhook_attempts`: Counter for delivery attempts
- Updated in database after webhook delivery

### 4. Retry Conditions

**Retries are attempted for:**
- Network errors (connection failures, timeouts)
- 5xx server errors (500-599)
- 429 Too Many Requests

**No retries for:**
- 4xx client errors (except 429)
- Invalid request format
- Context cancellation

### 5. Logging

Comprehensive logging at all stages:

- **INFO**: Webhook sending, successful delivery, retry attempts
- **WARN**: Failed delivery attempts, network errors
- **ERROR**: Final failure after all retries

## Configuration

### Environment Variables

Add to `.env` file:

```env
# Webhook Configuration
WEBHOOK_TIMEOUT=10s              # HTTP request timeout
WEBHOOK_MAX_RETRIES=3            # Maximum retry attempts
WEBHOOK_RETRY_DELAY=5s           # Initial retry delay
WEBHOOK_MAX_RETRY_DELAY=60s      # Maximum retry delay
```

### Config Structure

```go
type WebhookConfig struct {
    Timeout       time.Duration
    MaxRetries    int
    RetryDelay    time.Duration
    MaxRetryDelay time.Duration
}
```

## Usage

### Setting Up Webhook for a Session

When creating a WhatsApp session, provide webhook URL and optional secret:

```go
input := CreateSessionInput{
    SessionName:   "My Session",
    WebhookURL:    "https://example.com/webhook",
    WebhookSecret: "my-secret-key", // Optional
}

session, err := service.CreateSession(ctx, userID, input)
```

### Webhook Payload Format

Incoming messages trigger webhooks with this payload:

```json
{
  "event": "message.received",
  "session_id": "uuid-of-session",
  "timestamp": "2025-10-09T12:00:00Z",
  "data": {
    "message_id": "whatsapp-message-id",
    "from": "+1234567890",
    "to": "+0987654321",
    "content": "Message text content",
    "message_type": "text",
    "timestamp": "2025-10-09T12:00:00Z",
    "direction": "inbound"
  }
}
```

### Webhook Headers

The webhook client sends these headers:

```
Content-Type: application/json
User-Agent: WhatsApp-Webhook-Client/1.0
X-Webhook-Event: message.received
X-Webhook-Timestamp: 2025-10-09T12:00:00Z
X-Webhook-Signature: <hmac-sha256-hex> (if secret configured)
```

### Receiving Webhooks

Example webhook receiver implementation:

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    // Read body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    
    // Verify signature (if using HMAC)
    signature := r.Header.Get("X-Webhook-Signature")
    secret := "my-secret-key"
    if !webhook.VerifyWebhookSignature(body, signature, secret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Parse payload
    var payload webhook.WebhookPayload
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Process webhook
    // ... your business logic ...
    
    // Return success
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

## Error Handling

### Webhook Delivery Failures

When webhook delivery fails after all retries:

1. **Message Status**: Message remains in database with delivery status
2. **Webhook Attempts**: Counter shows number of attempts made
3. **Logging**: Error logged with full context
4. **Non-Blocking**: Message processing continues normally

### Best Practices

1. **Idempotency**: Webhook receivers should handle duplicate deliveries
2. **Fast Response**: Return 200 OK quickly, process asynchronously
3. **Signature Verification**: Always verify HMAC signature if using secrets
4. **Error Handling**: Return appropriate HTTP status codes
5. **Monitoring**: Monitor webhook delivery success rates

## Testing

### Unit Tests

Run webhook client tests:

```bash
go test -v ./internal/pkg/webhook/...
```

### Integration Testing

Test webhook delivery with a test server:

```go
// Create test server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Verify request
    // Return response
    w.WriteHeader(http.StatusOK)
}))
defer server.Close()

// Send webhook
req := webhook.WebhookRequest{
    URL: server.URL,
    Payload: webhook.WebhookPayload{
        Event:     "message.received",
        SessionID: "test-session",
        Timestamp: time.Now(),
        Data:      map[string]interface{}{},
    },
}

response := client.SendWebhook(ctx, req)
```

## Performance Considerations

### Asynchronous Delivery

Webhooks are delivered asynchronously using goroutines:

```go
if session.WebhookURL != "" && s.webhookClient != nil {
    go s.deliverWebhook(context.Background(), session, inboundMessage)
}
```

**Benefits:**
- Non-blocking message processing
- Fast message handling
- Independent webhook failures

**Considerations:**
- Webhook delivery happens in background
- Database updates may be delayed
- Monitor goroutine count in production

### Resource Usage

- **HTTP Connections**: Reuses HTTP client with connection pooling
- **Timeouts**: Configurable timeout prevents hanging requests
- **Memory**: Minimal overhead per webhook delivery
- **Goroutines**: One goroutine per webhook delivery

## Monitoring and Observability

### Metrics to Track

1. **Webhook Success Rate**: Percentage of successful deliveries
2. **Retry Rate**: Percentage of webhooks requiring retries
3. **Average Latency**: Time to deliver webhooks
4. **Failure Rate**: Webhooks failing after all retries

### Log Analysis

Search logs for webhook issues:

```bash
# Successful deliveries
grep "Webhook delivered successfully" logs.json

# Failed deliveries
grep "Webhook delivery failed after all retries" logs.json

# Retry attempts
grep "Retrying webhook after delay" logs.json
```

### Database Queries

Check webhook delivery status:

```sql
-- Messages with failed webhook delivery
SELECT * FROM whatsapp_messages 
WHERE webhook_delivered = false 
AND webhook_attempts >= 3;

-- Webhook success rate
SELECT 
    COUNT(*) FILTER (WHERE webhook_delivered = true) * 100.0 / COUNT(*) as success_rate
FROM whatsapp_messages 
WHERE direction = 'inbound';
```

## Security Considerations

### HMAC Signature Security

1. **Secret Management**: Store webhook secrets securely (encrypted in DB)
2. **Secret Rotation**: Support updating webhook secrets
3. **Signature Verification**: Always verify signatures on receiver side
4. **Timing Attacks**: Use constant-time comparison (`hmac.Equal`)

### Network Security

1. **HTTPS Only**: Recommend HTTPS for webhook URLs in production
2. **Timeouts**: Prevent resource exhaustion with timeouts
3. **Rate Limiting**: Consider rate limiting webhook deliveries
4. **IP Whitelisting**: Webhook receivers can whitelist source IPs

## Troubleshooting

### Common Issues

**Issue: Webhooks not being delivered**
- Check webhook URL is accessible
- Verify network connectivity
- Check logs for error messages
- Ensure webhook URL returns 2xx status

**Issue: Signature verification fails**
- Verify secret matches on both sides
- Check payload is not modified
- Ensure using raw body for verification
- Verify hex encoding of signature

**Issue: Too many retries**
- Check webhook endpoint performance
- Verify endpoint returns 2xx on success
- Consider increasing timeout
- Check for network issues

**Issue: Webhooks timing out**
- Increase `WEBHOOK_TIMEOUT` configuration
- Optimize webhook receiver performance
- Check network latency
- Consider async processing on receiver

## Requirements Satisfied

This implementation satisfies the following requirements:

- **4.3**: Webhook delivery when message is received
- **4.4**: Authentication headers (HMAC signature)
- **4.5**: Retry with exponential backoff (up to 3 attempts)
- **4.6**: Logging of failed webhooks after max retries
- **9.5**: Error handling and resilience

## Future Enhancements

Potential improvements for future iterations:

1. **Webhook Events**: Support more event types (connection, disconnection, etc.)
2. **Batch Delivery**: Batch multiple messages in single webhook
3. **Dead Letter Queue**: Store failed webhooks for manual retry
4. **Webhook Management API**: CRUD operations for webhook configuration
5. **Webhook Testing**: Test webhook endpoint before saving
6. **Circuit Breaker**: Temporarily disable failing webhooks
7. **Metrics Export**: Export webhook metrics to monitoring systems
