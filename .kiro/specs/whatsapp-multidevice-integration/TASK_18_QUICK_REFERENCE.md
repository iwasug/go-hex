# Task 18: Configuration Management - Quick Reference

## Configuration Structure

```go
type WhatsAppConfig struct {
    Enabled                bool          // Enable/disable WhatsApp
    QRTimeout              time.Duration // QR code expiration
    ReconnectInterval      time.Duration // Reconnection interval
    MaxReconnectAttempts   int          // Max reconnect attempts
    MessageRetentionDays   int          // Message retention period
    MediaMaxSizeMB         int          // Max media file size
    MediaStoragePath       string       // Media storage path
    WebhookTimeout         time.Duration // Webhook timeout
    WebhookMaxRetries      int          // Webhook max retries
    WebhookRetryDelay      time.Duration // Webhook retry delay
}
```

## Environment Variables

```bash
# WhatsApp Configuration
WHATSAPP_ENABLED=true
WHATSAPP_QR_TIMEOUT=60s
WHATSAPP_RECONNECT_INTERVAL=30s
WHATSAPP_MAX_RECONNECT_ATTEMPTS=5
WHATSAPP_MESSAGE_RETENTION_DAYS=90
WHATSAPP_MEDIA_MAX_SIZE_MB=16
WHATSAPP_MEDIA_STORAGE_PATH=./data/whatsapp/media
WHATSAPP_WEBHOOK_TIMEOUT=10s
WHATSAPP_WEBHOOK_MAX_RETRIES=3
WHATSAPP_WEBHOOK_RETRY_DELAY=5s
```

## Usage

```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Access WhatsApp config
if cfg.WhatsApp.Enabled {
    qrTimeout := cfg.WhatsApp.QRTimeout
    mediaPath := cfg.WhatsApp.MediaStoragePath
    maxSize := cfg.WhatsApp.MediaMaxSizeMB
}
```

## Validation Rules

When `WHATSAPP_ENABLED=true`:
- ✅ QRTimeout > 0
- ✅ ReconnectInterval > 0
- ✅ MaxReconnectAttempts >= 0
- ✅ MessageRetentionDays >= 0
- ✅ MediaMaxSizeMB > 0
- ✅ MediaStoragePath != ""
- ✅ WebhookTimeout > 0
- ✅ WebhookMaxRetries >= 0
- ✅ WebhookRetryDelay > 0

When `WHATSAPP_ENABLED=false`:
- ⏭️ All validation is skipped

## Default Values

| Setting | Default |
|---------|---------|
| Enabled | `true` |
| QRTimeout | `60s` |
| ReconnectInterval | `30s` |
| MaxReconnectAttempts | `5` |
| MessageRetentionDays | `90` |
| MediaMaxSizeMB | `16` |
| MediaStoragePath | `./data/whatsapp/media` |
| WebhookTimeout | `10s` |
| WebhookMaxRetries | `3` |
| WebhookRetryDelay | `5s` |

## Testing

Run tests:
```bash
go test ./internal/config/... -v
```

Test specific function:
```bash
go test ./internal/config/... -v -run TestWhatsAppConfig
```
