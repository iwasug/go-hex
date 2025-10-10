package repository

import (
	"context"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// ListMessagesParams contains parameters for listing WhatsApp messages with filtering and pagination
type ListMessagesParams struct {
	SessionID   string                   // Filter by session ID (required)
	Direction   entity.MessageDirection  // Filter by message direction (empty for all)
	Status      entity.MessageStatus     // Filter by message status (empty for all)
	MessageType string                   // Filter by message type (e.g., "text", "image", "video")
	StartDate   *time.Time               // Filter messages from this date (inclusive)
	EndDate     *time.Time               // Filter messages until this date (inclusive)
	Search      string                   // Search in message content
	OrderBy     string                   // Field to order by (e.g., "timestamp", "created_at")
	Order       string                   // "asc" or "desc"
	Limit       int                      // Number of records to return
	Offset      int                      // Number of records to skip
}

// WhatsAppMessageRepository defines the interface for WhatsApp message data operations
type WhatsAppMessageRepository interface {
	// Create creates a new WhatsApp message in the repository
	Create(ctx context.Context, message *entity.WhatsAppMessage) error

	// FindByID retrieves a WhatsApp message by ID
	FindByID(ctx context.Context, id string) (*entity.WhatsAppMessage, error)

	// FindByMessageID retrieves a WhatsApp message by WhatsApp message ID
	FindByMessageID(ctx context.Context, messageID string) (*entity.WhatsAppMessage, error)

	// FindBySessionID retrieves all messages for a specific session
	FindBySessionID(ctx context.Context, sessionID string) ([]*entity.WhatsAppMessage, error)

	// Update updates an existing WhatsApp message
	Update(ctx context.Context, message *entity.WhatsAppMessage) error

	// Delete deletes a WhatsApp message from the repository
	Delete(ctx context.Context, id string) error

	// List retrieves a list of WhatsApp messages with filtering, sorting, and pagination
	// Returns the list of messages, total count, and error
	List(ctx context.Context, params ListMessagesParams) ([]*entity.WhatsAppMessage, int, error)

	// UpdateStatus updates the status of a WhatsApp message
	UpdateStatus(ctx context.Context, id string, status entity.MessageStatus) error

	// IncrementWebhookAttempts increments the webhook delivery attempt counter
	IncrementWebhookAttempts(ctx context.Context, id string) error

	// MarkWebhookDelivered marks a message's webhook as successfully delivered
	MarkWebhookDelivered(ctx context.Context, id string) error

	// FindPendingWebhooks retrieves messages that need webhook delivery
	// Returns messages where webhook is not delivered and attempts are below max
	FindPendingWebhooks(ctx context.Context, maxAttempts int, limit int) ([]*entity.WhatsAppMessage, error)

	// FindBySessionIDAndDirection retrieves messages filtered by session and direction
	FindBySessionIDAndDirection(ctx context.Context, sessionID string, direction entity.MessageDirection) ([]*entity.WhatsAppMessage, error)

	// CountBySessionID counts the number of messages for a specific session
	CountBySessionID(ctx context.Context, sessionID string) (int, error)

	// DeleteBySessionID deletes all messages for a specific session
	DeleteBySessionID(ctx context.Context, sessionID string) error

	// FindRecentBySessionID retrieves the most recent messages for a session
	FindRecentBySessionID(ctx context.Context, sessionID string, limit int) ([]*entity.WhatsAppMessage, error)
}
