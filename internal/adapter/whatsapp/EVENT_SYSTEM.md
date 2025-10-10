# WhatsApp Event Handling System

## Overview

The WhatsApp event handling system provides a robust, scalable architecture for processing WhatsApp events such as connections, disconnections, QR codes, and incoming messages. The system is built around an event bus pattern that decouples event generation from event processing.

## Architecture

### Components

1. **EventHandler** (`event_handler.go`)
   - Receives raw events from whatsmeow library
   - Dispatches events to appropriate handlers
   - Transforms whatsmeow events into domain events

2. **EventBus** (`event_bus.go`)
   - Central event distribution system
   - Manages event subscribers
   - Handles database updates for events
   - Implements reconnection logic

3. **EventHandlers** (callbacks)
   - Simple callback functions registered with EventHandler
   - Bridge between EventHandler and EventBus
   - Defined in ClientManager

4. **ClientManager** (`client_manager.go`)
   - Creates and manages WhatsApp clients
   - Initializes event bus and handlers
   - Provides subscription interface

## Event Flow

```
WhatsApp Server
      ↓
whatsmeow.Client
      ↓
EventHandler.HandleEvent()
      ↓
EventHandlers callbacks (OnConnected, OnDisconnected, etc.)
      ↓
EventBus.Handle*Event()
      ↓
├─ Update Database (session status, messages, etc.)
├─ Publish to Subscribers
└─ Trigger side effects (reconnection, webhooks, etc.)
```

## Event Types

### 1. Connected Event

**Triggered when:** WhatsApp client successfully connects to WhatsApp servers

**Actions:**
- Update session status to `connected`
- Update phone number (if available)
- Clear QR code data
- Update last activity timestamp
- Publish to subscribers

**Event Data:**
```go
type ConnectedEventData struct {
    Phone string
}
```

### 2. Disconnected Event

**Triggered when:** WhatsApp client loses connection or is logged out

**Actions:**
- Update session status to `disconnected`
- Update last activity timestamp
- Publish to subscribers
- **Initiate automatic reconnection** (async)

**Event Data:**
```go
type DisconnectedEventData struct {
    Reason string
}
```

**Reconnection Logic:**
- Maximum 5 reconnection attempts
- Exponential backoff: 5s, 10s, 20s, 40s, 80s (capped at 5 minutes)
- Updates session status to `connecting` during attempts
- Updates session status to `failed` if all attempts fail
- Stops if client is already reconnected or doesn't exist

### 3. QR Code Event

**Triggered when:** New QR code is generated for device pairing

**Actions:**
- Store QR code in database
- Set expiry time (60 seconds)
- Update session status to `authenticating`
- Update last activity timestamp
- Publish to subscribers

**Event Data:**
```go
type QRCodeEventData struct {
    QRCode string
    Expiry time.Time
}
```

### 4. Message Event

**Triggered when:** Incoming message is received from WhatsApp

**Actions:**
- Create message entity
- Store message in database
- Update session last activity
- Publish to subscribers (for webhook delivery)

**Event Data:**
```go
type MessageEventData struct {
    Message IncomingMessage
}

type IncomingMessage struct {
    MessageID   string
    From        string
    To          string
    Content     string
    MessageType string // text, image, video, audio, document
    MediaURL    string
    Caption     string
    Timestamp   time.Time
}
```

**Supported Message Types:**
- Text messages (Conversation, ExtendedTextMessage)
- Image messages (with optional caption)
- Video messages (with optional caption)
- Audio messages
- Document messages (with filename)

### 5. Pair Success Event

**Triggered when:** Device pairing is successfully completed

**Actions:**
- Update session status to `connected`
- Update phone number
- Clear QR code data
- Update last activity timestamp
- Publish to subscribers

**Event Data:**
```go
type PairSuccessEventData struct {
    Phone string
}
```

## Usage

### Basic Setup

```go
// Create repositories
sessionRepo := postgres.NewWhatsAppSessionRepository(db, logger)
messageRepo := postgres.NewWhatsAppMessageRepository(db, logger)

// Create client manager (event bus is created internally)
clientManager := whatsapp.NewClientManager(
    db,
    sessionRepo,
    messageRepo,
    logger,
)

// Create a client (event handlers are registered automatically)
err := clientManager.CreateClient(ctx, sessionID, userID)
```

### Subscribing to Events

```go
// Subscribe to message events for webhook delivery
clientManager.SubscribeToEvent(whatsapp.EventTypeMessage, func(ctx context.Context, event whatsapp.Event) error {
    data := event.Data.(whatsapp.MessageEventData)
    
    // Get session to check webhook URL
    session, err := sessionRepo.FindByID(ctx, event.SessionID)
    if err != nil {
        return err
    }
    
    if session.WebhookURL != "" {
        // Deliver webhook
        return webhookClient.DeliverMessage(ctx, session.WebhookURL, data.Message)
    }
    
    return nil
})

// Subscribe to connected events for notifications
clientManager.SubscribeToEvent(whatsapp.EventTypeConnected, func(ctx context.Context, event whatsapp.Event) error {
    data := event.Data.(whatsapp.ConnectedEventData)
    
    // Send notification to user
    return notificationService.NotifySessionConnected(ctx, event.UserID, data.Phone)
})
```

### Event Types Available for Subscription

- `EventTypeConnected` - Session connected
- `EventTypeDisconnected` - Session disconnected
- `EventTypeQRCode` - QR code generated
- `EventTypeMessage` - Message received
- `EventTypePairSuccess` - Pairing successful

## Automatic Reconnection

The event bus implements automatic reconnection for disconnected sessions:

### Configuration

```go
const (
    maxAttempts       = 5                // Maximum reconnection attempts
    initialDelay      = 5 * time.Second  // Initial delay before first attempt
    maxDelay          = 5 * time.Minute  // Maximum delay between attempts
    backoffMultiplier = 2                // Exponential backoff multiplier
)
```

### Reconnection Flow

1. Session disconnects → `OnDisconnected` event triggered
2. Event bus updates session status to `disconnected`
3. Reconnection goroutine starts (async)
4. For each attempt:
   - Wait with exponential backoff
   - Update status to `connecting`
   - Get client from manager
   - Check if already connected (skip if yes)
   - Attempt to connect
   - If successful: update status to `connected` and return
   - If failed: increase delay and continue
5. If all attempts fail: update status to `failed`

### Reconnection Delays

- Attempt 1: 5 seconds
- Attempt 2: 10 seconds
- Attempt 3: 20 seconds
- Attempt 4: 40 seconds
- Attempt 5: 80 seconds

## Error Handling

### Database Errors

- Critical errors (status updates, message storage) are logged and returned
- Non-critical errors (last activity updates) are logged but don't stop processing
- Failed database operations don't prevent event publishing to subscribers

### Subscriber Errors

- Subscribers are called asynchronously (non-blocking)
- Errors in subscribers are logged but don't affect other subscribers
- Each subscriber runs in its own goroutine

### Reconnection Errors

- Connection failures trigger exponential backoff
- Missing clients stop reconnection attempts
- All failures are logged with context

## Thread Safety

- EventBus uses `sync.RWMutex` for subscriber management
- Subscribers are called in separate goroutines
- ClientManager ensures thread-safe client access
- Event publishing is non-blocking

## Testing

### Unit Testing Event Handlers

```go
func TestEventBus_HandleConnectedEvent(t *testing.T) {
    // Create mock repositories
    sessionRepo := &mockSessionRepo{}
    messageRepo := &mockMessageRepo{}
    logger := &mockLogger{}
    
    // Create event bus
    eventBus := whatsapp.NewEventBus(sessionRepo, messageRepo, logger)
    
    // Test connected event
    err := eventBus.HandleConnectedEvent(ctx, "session-123", "user-456", "+1234567890")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, entity.SessionStatusConnected, sessionRepo.lastStatus)
    assert.Equal(t, "+1234567890", sessionRepo.lastPhone)
}
```

### Integration Testing

```go
func TestEventFlow_MessageReceived(t *testing.T) {
    // Setup real database and repositories
    db := setupTestDB(t)
    sessionRepo := postgres.NewWhatsAppSessionRepository(db, logger)
    messageRepo := postgres.NewWhatsAppMessageRepository(db, logger)
    
    // Create client manager
    clientManager := whatsapp.NewClientManager(db, sessionRepo, messageRepo, logger)
    
    // Subscribe to message events
    var receivedEvent whatsapp.Event
    clientManager.SubscribeToEvent(whatsapp.EventTypeMessage, func(ctx context.Context, event whatsapp.Event) error {
        receivedEvent = event
        return nil
    })
    
    // Simulate message event
    message := whatsapp.IncomingMessage{
        MessageID: "msg-123",
        From: "+1234567890",
        Content: "Hello",
        MessageType: "text",
        Timestamp: time.Now(),
    }
    
    err := clientManager.GetEventBus().HandleMessageEvent(ctx, "session-123", "user-456", message)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, receivedEvent)
    
    // Verify message in database
    storedMsg, err := messageRepo.FindByMessageID(ctx, "msg-123")
    assert.NoError(t, err)
    assert.Equal(t, "Hello", storedMsg.Content)
}
```

## Best Practices

1. **Always handle errors in subscribers** - Don't let subscriber errors crash the application
2. **Keep subscribers lightweight** - Heavy processing should be done asynchronously
3. **Use context for cancellation** - Pass context to long-running operations
4. **Log all events** - Helps with debugging and monitoring
5. **Test reconnection logic** - Ensure sessions recover from disconnections
6. **Monitor event bus performance** - Track subscriber execution times
7. **Implement idempotency** - Handle duplicate events gracefully

## Monitoring

### Key Metrics to Track

- Event processing time per type
- Subscriber execution time
- Reconnection success rate
- Failed reconnection count
- Event queue depth (if implementing queuing)
- Database operation latency

### Logging

All events are logged with structured fields:
- `event_type` - Type of event
- `session_id` - Session identifier
- `user_id` - User identifier
- `error` - Error message (if any)
- Additional context-specific fields

## Future Enhancements

1. **Event Persistence** - Store events in database for replay
2. **Event Queuing** - Use message queue for high-volume scenarios
3. **Circuit Breaker** - Prevent cascading failures in subscribers
4. **Event Filtering** - Allow subscribers to filter events by criteria
5. **Event Replay** - Replay events for debugging or recovery
6. **Metrics Collection** - Built-in Prometheus metrics
7. **Dead Letter Queue** - Handle permanently failed events

## Requirements Satisfied

This implementation satisfies the following requirements from the spec:

- **Requirement 2.3**: QR code expiry and automatic generation
- **Requirement 2.4**: Authentication success handling
- **Requirement 2.5**: Session status updates
- **Requirement 4.1**: Message capture and storage
- **Requirement 4.2**: Message persistence
- **Requirement 5.2**: Automatic reconnection
- **Requirement 5.4**: Last activity tracking
