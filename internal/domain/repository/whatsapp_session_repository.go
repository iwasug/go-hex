package repository

import (
	"context"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// ListSessionsParams contains parameters for listing WhatsApp sessions with filtering and pagination
type ListSessionsParams struct {
	UserID     string                // Filter by user ID (empty for all users - admin only)
	Status     entity.SessionStatus  // Filter by session status (empty for all statuses)
	IsActive   *bool                 // Filter by active status (nil for all)
	Search     string                // Search by session name or phone number
	OrderBy    string                // Field to order by (e.g., "session_name", "created_at", "last_activity")
	Order      string                // "asc" or "desc"
	Limit      int                   // Number of records to return
	Offset     int                   // Number of records to skip
	IncludeAll bool                  // Admin: view all sessions across all users
}

// WhatsAppSessionRepository defines the interface for WhatsApp session data operations
type WhatsAppSessionRepository interface {
	// Create creates a new WhatsApp session in the repository
	Create(ctx context.Context, session *entity.WhatsAppSession) error

	// FindByID retrieves a WhatsApp session by ID
	FindByID(ctx context.Context, id string) (*entity.WhatsAppSession, error)

	// FindByUserID retrieves all WhatsApp sessions for a specific user
	FindByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error)

	// FindByUserIDAndName retrieves a session by user ID and session name
	FindByUserIDAndName(ctx context.Context, userID, sessionName string) (*entity.WhatsAppSession, error)

	// Update updates an existing WhatsApp session
	Update(ctx context.Context, session *entity.WhatsAppSession) error

	// Delete deletes a WhatsApp session from the repository
	Delete(ctx context.Context, id string) error

	// List retrieves a list of WhatsApp sessions with filtering, sorting, and pagination
	// Returns the list of sessions, total count, and error
	List(ctx context.Context, params ListSessionsParams) ([]*entity.WhatsAppSession, int, error)

	// UpdateStatus updates the status of a WhatsApp session
	UpdateStatus(ctx context.Context, id string, status entity.SessionStatus) error

	// UpdateQRCode updates the QR code and expiry time for a session
	UpdateQRCode(ctx context.Context, id string, qrCode string, expiry time.Time) error

	// ClearQRCode clears the QR code data for a session
	ClearQRCode(ctx context.Context, id string) error

	// UpdatePhoneNumber updates the connected phone number for a session
	UpdatePhoneNumber(ctx context.Context, id string, phoneNumber string) error

	// UpdateLastActivity updates the last activity timestamp for a session
	UpdateLastActivity(ctx context.Context, id string) error

	// UpdateDeviceInfo updates the device information for a session
	UpdateDeviceInfo(ctx context.Context, id string, deviceInfo string) error

	// FindActiveByUserID retrieves all active WhatsApp sessions for a specific user
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error)

	// FindConnectedSessions retrieves all sessions with connected status
	FindConnectedSessions(ctx context.Context) ([]*entity.WhatsAppSession, error)

	// CountByUserID counts the number of sessions for a specific user
	CountByUserID(ctx context.Context, userID string) (int, error)

	// IncrementReconnectAttempts increments the reconnection attempt counter
	IncrementReconnectAttempts(ctx context.Context, id string) error

	// ResetReconnectAttempts resets the reconnection attempt counter
	ResetReconnectAttempts(ctx context.Context, id string) error
}
