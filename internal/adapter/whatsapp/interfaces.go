package whatsapp

import (
	"context"
	"time"
)

// ClientManagerInterface defines the interface for managing WhatsApp clients
type ClientManagerInterface interface {
	// CreateClient creates a new WhatsApp client for a session
	CreateClient(ctx context.Context, sessionID, userID string) error

	// GetClient retrieves a WhatsApp client by session ID (thread-safe)
	GetClient(sessionID string) (*Client, error)

	// DeleteClient removes a WhatsApp client and performs cleanup
	DeleteClient(ctx context.Context, sessionID string) error

	// RestoreClients restores all connected sessions on application startup
	RestoreClients(ctx context.Context) error

	// GetAllClients returns all active clients (for monitoring/debugging)
	GetAllClients() map[string]*Client

	// GetClientCount returns the number of active clients
	GetClientCount() int

	// Shutdown gracefully shuts down all clients
	Shutdown(ctx context.Context) error

	// GetEventBus returns the event bus instance
	GetEventBus() *EventBus

	// SubscribeToEvent subscribes to a specific event type
	SubscribeToEvent(eventType EventType, subscriber EventSubscriber)
}

// ClientInterface defines the interface for a WhatsApp client
type ClientInterface interface {
	// Connect establishes a WebSocket connection to WhatsApp
	Connect() error

	// Disconnect gracefully disconnects from WhatsApp
	Disconnect() error

	// GetQRCode generates and returns a QR code as base64-encoded PNG
	GetQRCode(ctx context.Context, timeout time.Duration) (string, error)

	// SendTextMessage sends a text message to a recipient
	SendTextMessage(ctx context.Context, to, content string) (string, error)

	// SendMediaMessage sends a media message to a recipient
	SendMediaMessage(ctx context.Context, to string, media MediaMessage) (string, error)

	// IsConnected returns whether the client is connected
	IsConnected() bool

	// IsLoggedIn returns whether the client is logged in
	IsLoggedIn() bool

	// GetConnectedPhone returns the connected phone number
	GetConnectedPhone() string

	// GetSessionID returns the session ID
	GetSessionID() string

	// GetUserID returns the user ID
	GetUserID() string

	// Logout logs out from WhatsApp and removes device data
	Logout(ctx context.Context) error
}
