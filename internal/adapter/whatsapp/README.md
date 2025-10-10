# WhatsApp Client Manager

This package provides a WhatsApp client manager implementation using the whatsmeow library with PostgreSQL storage.

## Components

### ClientManager

The `ClientManager` manages the lifecycle of multiple WhatsApp clients, providing thread-safe operations for creating, retrieving, and deleting clients.

**Key Features:**
- Thread-safe client management using RWMutex
- PostgreSQL-backed storage for session persistence
- Automatic client restoration on application startup
- Graceful shutdown of all clients
- Connection pool management

**Methods:**
- `CreateClient(ctx, sessionID, userID)` - Creates a new WhatsApp client
- `GetClient(sessionID)` - Retrieves an existing client (thread-safe)
- `DeleteClient(ctx, sessionID)` - Removes a client with proper cleanup
- `RestoreClients(ctx)` - Restores all connected sessions on startup
- `GetAllClients()` - Returns all active clients
- `GetClientCount()` - Returns the number of active clients
- `Shutdown(ctx)` - Gracefully shuts down all clients

### Client

The `Client` wraps a whatsmeow.Client and provides additional functionality for message sending, QR code generation, and connection management.

**Key Features:**
- WebSocket connection management
- QR code generation for device pairing
- Text and media message sending
- Event handling for incoming messages
- Thread-safe operations

**Methods:**
- `Connect()` - Establishes WebSocket connection
- `Disconnect()` - Gracefully disconnects
- `GetQRCode(ctx, timeout)` - Generates QR code for pairing
- `SendTextMessage(ctx, to, content)` - Sends text message
- `SendMediaMessage(ctx, to, media)` - Sends media message
- `IsConnected()` - Checks connection status
- `IsLoggedIn()` - Checks login status
- `GetConnectedPhone()` - Returns connected phone number
- `Logout()` - Logs out and removes device data

### EventHandler

The `EventHandler` processes events from the WhatsApp client and dispatches them to registered callbacks.

**Supported Events:**
- `OnConnected` - Fired when client connects
- `OnDisconnected` - Fired when client disconnects
- `OnQRCode` - Fired when QR code is generated
- `OnMessage` - Fired when message is received
- `OnPairSuccess` - Fired when pairing is successful

### PostgresStore

The `PostgresStore` implements whatsmeow's store interfaces using PostgreSQL, enabling:
- True horizontal scaling in Kubernetes
- No file storage dependencies
- High availability with PostgreSQL clustering
- Standard backup/restore procedures

## Usage Example

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    "golang-chi-hexagonal-base/internal/adapter/whatsapp"
    "golang-chi-hexagonal-base/internal/pkg/logger"
)

func main() {
    // Initialize dependencies
    db := initDatabase()
    sessionRepo := initSessionRepository(db)
    log := logger.NewFromConfig("info", "json")
    
    // Define event handlers
    eventHandlers := whatsapp.EventHandlers{
        OnConnected: func(sessionID, userID, phone string) {
            log.Info("Session connected", logger.Fields{
                "session_id": sessionID,
                "phone": phone,
            })
        },
        OnDisconnected: func(sessionID, userID string) {
            log.Warn("Session disconnected", logger.Fields{
                "session_id": sessionID,
            })
        },
        OnMessage: func(sessionID, userID string, msg whatsapp.IncomingMessage) {
            log.Info("Message received", logger.Fields{
                "session_id": sessionID,
                "from": msg.From,
                "content": msg.Content,
            })
        },
    }
    
    // Create client manager
    clientManager := whatsapp.NewClientManager(
        db,
        sessionRepo,
        log,
        eventHandlers,
    )
    
    // Restore clients on startup
    ctx := context.Background()
    if err := clientManager.RestoreClients(ctx); err != nil {
        log.Error("Failed to restore clients", logger.Fields{
            "error": err.Error(),
        })
    }
    
    // Create a new client
    sessionID := "session-123"
    userID := "user-456"
    
    if err := clientManager.CreateClient(ctx, sessionID, userID); err != nil {
        log.Fatal("Failed to create client", logger.Fields{
            "error": err.Error(),
        })
    }
    
    // Get the client
    client, err := clientManager.GetClient(sessionID)
    if err != nil {
        log.Fatal("Failed to get client", logger.Fields{
            "error": err.Error(),
        })
    }
    
    // Connect the client
    if err := client.Connect(); err != nil {
        log.Fatal("Failed to connect", logger.Fields{
            "error": err.Error(),
        })
    }
    
    // Send a message
    messageID, err := client.SendTextMessage(ctx, "+1234567890", "Hello, World!")
    if err != nil {
        log.Error("Failed to send message", logger.Fields{
            "error": err.Error(),
        })
    } else {
        log.Info("Message sent", logger.Fields{
            "message_id": messageID,
        })
    }
    
    // Graceful shutdown
    defer func() {
        if err := clientManager.Shutdown(ctx); err != nil {
            log.Error("Failed to shutdown", logger.Fields{
                "error": err.Error(),
            })
        }
    }()
}
```

## Media Message Example

```go
// Send an image
media := whatsapp.MediaMessage{
    Type:     whatsapp.MediaTypeImage,
    Data:     imageBytes,
    MimeType: "image/jpeg",
    Caption:  "Check out this image!",
}

messageID, err := client.SendMediaMessage(ctx, "+1234567890", media)
if err != nil {
    log.Error("Failed to send image", logger.Fields{
        "error": err.Error(),
    })
}
```

## QR Code Generation Example

```go
import "time"

// Generate QR code with 60 second timeout
qrCode, err := client.GetQRCode(ctx, 60*time.Second)
if err != nil {
    log.Error("Failed to get QR code", logger.Fields{
        "error": err.Error(),
    })
    return
}

// qrCode is a base64-encoded PNG image
// Display it to the user for scanning
```

## Thread Safety

All operations on `ClientManager` and `Client` are thread-safe:
- `ClientManager` uses `sync.RWMutex` for concurrent access to the client map
- `Client` uses `sync.RWMutex` for concurrent access to connection state
- Multiple goroutines can safely call methods on the same instances

## Error Handling

All methods return descriptive errors:
- Connection errors include context about the failure
- Invalid operations (e.g., sending when not logged in) return clear error messages
- All errors are logged with relevant context fields

## Logging

The package uses structured logging with the following fields:
- `component`: Always set to "client_manager" or "client"
- `session_id`: The session identifier
- `user_id`: The user identifier
- `error`: Error message when applicable
- Additional context-specific fields

## Requirements

This implementation satisfies the following requirements:
- 1.1: WhatsApp session management
- 1.2: Session data persistence using PostgreSQL
- 1.6: Automatic session restoration on application restart
- 5.2: Automatic reconnection on disconnect
- 5.5: Session health monitoring
- 9.5: Error handling with retry logic
