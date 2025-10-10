package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

// WhatsAppMessageRepository implements the repository.WhatsAppMessageRepository interface using PostgreSQL
type WhatsAppMessageRepository struct {
	db *DB
}

// NewWhatsAppMessageRepository creates a new WhatsAppMessageRepository instance
func NewWhatsAppMessageRepository(db *DB) *WhatsAppMessageRepository {
	return &WhatsAppMessageRepository{db: db}
}

// Create creates a new WhatsApp message in the database
func (r *WhatsAppMessageRepository) Create(ctx context.Context, message *entity.WhatsAppMessage) error {
	query := `
		INSERT INTO whatsapp_messages (
			id, session_id, message_id, direction, status, from_number, to_number,
			content, message_type, timestamp, webhook_delivered, webhook_attempts,
			error_message, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.SessionID,
		message.MessageID,
		message.Direction,
		message.Status,
		message.From,
		message.To,
		message.Content,
		message.MessageType,
		message.Timestamp,
		message.WebhookDelivered,
		message.WebhookAttempts,
		message.ErrorMessage,
		message.CreatedAt,
		message.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("message with this ID already exists for session: %w", err)
		}
		if strings.Contains(err.Error(), "foreign key") {
			return fmt.Errorf("session not found: %w", err)
		}
		return fmt.Errorf("failed to create whatsapp message: %w", err)
	}

	return nil
}

// FindByID retrieves a WhatsApp message by ID
func (r *WhatsAppMessageRepository) FindByID(ctx context.Context, id string) (*entity.WhatsAppMessage, error) {
	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE id = $1
	`

	message := &entity.WhatsAppMessage{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.MessageID,
		&message.Direction,
		&message.Status,
		&message.From,
		&message.To,
		&message.Content,
		&message.MessageType,
		&message.Timestamp,
		&message.WebhookDelivered,
		&message.WebhookAttempts,
		&message.ErrorMessage,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("whatsapp message not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find whatsapp message by ID: %w", err)
	}

	return message, nil
}

// FindByMessageID retrieves a WhatsApp message by WhatsApp message ID
func (r *WhatsAppMessageRepository) FindByMessageID(ctx context.Context, messageID string) (*entity.WhatsAppMessage, error) {
	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE message_id = $1
	`

	message := &entity.WhatsAppMessage{}
	err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&message.ID,
		&message.SessionID,
		&message.MessageID,
		&message.Direction,
		&message.Status,
		&message.From,
		&message.To,
		&message.Content,
		&message.MessageType,
		&message.Timestamp,
		&message.WebhookDelivered,
		&message.WebhookAttempts,
		&message.ErrorMessage,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("whatsapp message not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find whatsapp message by message ID: %w", err)
	}

	return message, nil
}

// FindBySessionID retrieves all messages for a specific session
func (r *WhatsAppMessageRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*entity.WhatsAppMessage, error) {
	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE session_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find messages by session ID: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// Update updates an existing WhatsApp message
func (r *WhatsAppMessageRepository) Update(ctx context.Context, message *entity.WhatsAppMessage) error {
	query := `
		UPDATE whatsapp_messages
		SET message_id = $1, direction = $2, status = $3, from_number = $4, to_number = $5,
		    content = $6, message_type = $7, timestamp = $8, webhook_delivered = $9,
		    webhook_attempts = $10, error_message = $11, updated_at = $12
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		message.MessageID,
		message.Direction,
		message.Status,
		message.From,
		message.To,
		message.Content,
		message.MessageType,
		message.Timestamp,
		message.WebhookDelivered,
		message.WebhookAttempts,
		message.ErrorMessage,
		message.UpdatedAt,
		message.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update whatsapp message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp message not found")
	}

	return nil
}

// Delete deletes a WhatsApp message from the repository
func (r *WhatsAppMessageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM whatsapp_messages WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete whatsapp message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp message not found")
	}

	return nil
}

// List retrieves a list of WhatsApp messages with filtering, sorting, and pagination
func (r *WhatsAppMessageRepository) List(ctx context.Context, params repository.ListMessagesParams) ([]*entity.WhatsAppMessage, int, error) {
	// Validate order_by field (whitelist)
	validOrderByFields := map[string]bool{
		"timestamp":    true,
		"created_at":   true,
		"status":       true,
		"message_type": true,
	}

	orderBy := "timestamp"
	if params.OrderBy != "" {
		if !validOrderByFields[params.OrderBy] {
			return nil, 0, fmt.Errorf("invalid order_by field: %s", params.OrderBy)
		}
		orderBy = params.OrderBy
	}

	// Validate order direction
	order := "desc"
	if params.Order != "" {
		if params.Order != "asc" && params.Order != "desc" {
			return nil, 0, fmt.Errorf("invalid order direction: %s", params.Order)
		}
		order = params.Order
	}

	// Set default pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query with dynamic filtering
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT 
			id, session_id, message_id, direction, status, from_number, to_number,
			content, message_type, timestamp, webhook_delivered, webhook_attempts,
			error_message, created_at, updated_at,
			COUNT(*) OVER() as total_count
		FROM whatsapp_messages
		WHERE 1=1
	`)

	args := []interface{}{}
	argPos := 1

	// Filter by session ID (required)
	if params.SessionID != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND session_id = $%d", argPos))
		args = append(args, params.SessionID)
		argPos++
	}

	// Filter by direction
	if params.Direction != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND direction = $%d", argPos))
		args = append(args, params.Direction)
		argPos++
	}

	// Filter by status
	if params.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argPos))
		args = append(args, params.Status)
		argPos++
	}

	// Filter by message type
	if params.MessageType != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND message_type = $%d", argPos))
		args = append(args, params.MessageType)
		argPos++
	}

	// Filter by date range
	if params.StartDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND timestamp >= $%d", argPos))
		args = append(args, *params.StartDate)
		argPos++
	}

	if params.EndDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND timestamp <= $%d", argPos))
		args = append(args, *params.EndDate)
		argPos++
	}

	// Search in message content
	if params.Search != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND content ILIKE $%d", argPos))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern)
		argPos++
	}

	// Add ORDER BY clause
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", orderBy, order))

	// Add pagination
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list whatsapp messages: %w", err)
	}
	defer rows.Close()

	var messages []*entity.WhatsAppMessage
	var totalCount int

	for rows.Next() {
		message := &entity.WhatsAppMessage{}
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.MessageID,
			&message.Direction,
			&message.Status,
			&message.From,
			&message.To,
			&message.Content,
			&message.MessageType,
			&message.Timestamp,
			&message.WebhookDelivered,
			&message.WebhookAttempts,
			&message.ErrorMessage,
			&message.CreatedAt,
			&message.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan whatsapp message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating whatsapp messages: %w", err)
	}

	return messages, totalCount, nil
}

// UpdateStatus updates the status of a WhatsApp message
func (r *WhatsAppMessageRepository) UpdateStatus(ctx context.Context, id string, status entity.MessageStatus) error {
	query := `
		UPDATE whatsapp_messages
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp message not found")
	}

	return nil
}

// IncrementWebhookAttempts increments the webhook delivery attempt counter
func (r *WhatsAppMessageRepository) IncrementWebhookAttempts(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_messages
		SET webhook_attempts = webhook_attempts + 1, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to increment webhook attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp message not found")
	}

	return nil
}

// MarkWebhookDelivered marks a message's webhook as successfully delivered
func (r *WhatsAppMessageRepository) MarkWebhookDelivered(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_messages
		SET webhook_delivered = true, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark webhook as delivered: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp message not found")
	}

	return nil
}

// FindPendingWebhooks retrieves messages that need webhook delivery
func (r *WhatsAppMessageRepository) FindPendingWebhooks(ctx context.Context, maxAttempts int, limit int) ([]*entity.WhatsAppMessage, error) {
	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE webhook_delivered = false AND webhook_attempts < $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, maxAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending webhooks: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// FindBySessionIDAndDirection retrieves messages filtered by session and direction
func (r *WhatsAppMessageRepository) FindBySessionIDAndDirection(ctx context.Context, sessionID string, direction entity.MessageDirection) ([]*entity.WhatsAppMessage, error) {
	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE session_id = $1 AND direction = $2
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to find messages by session and direction: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// CountBySessionID counts the number of messages for a specific session
func (r *WhatsAppMessageRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	query := `SELECT COUNT(*) FROM whatsapp_messages WHERE session_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by session ID: %w", err)
	}

	return count, nil
}

// DeleteBySessionID deletes all messages for a specific session
func (r *WhatsAppMessageRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {
	query := `DELETE FROM whatsapp_messages WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete messages by session ID: %w", err)
	}

	return nil
}

// FindRecentBySessionID retrieves the most recent messages for a session
func (r *WhatsAppMessageRepository) FindRecentBySessionID(ctx context.Context, sessionID string, limit int) ([]*entity.WhatsAppMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, session_id, message_id, direction, status, from_number, to_number,
		       content, message_type, timestamp, webhook_delivered, webhook_attempts,
		       error_message, created_at, updated_at
		FROM whatsapp_messages
		WHERE session_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// scanMessages is a helper function to scan multiple messages from rows
func (r *WhatsAppMessageRepository) scanMessages(rows *sql.Rows) ([]*entity.WhatsAppMessage, error) {
	var messages []*entity.WhatsAppMessage

	for rows.Next() {
		message := &entity.WhatsAppMessage{}
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.MessageID,
			&message.Direction,
			&message.Status,
			&message.From,
			&message.To,
			&message.Content,
			&message.MessageType,
			&message.Timestamp,
			&message.WebhookDelivered,
			&message.WebhookAttempts,
			&message.ErrorMessage,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan whatsapp message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating whatsapp messages: %w", err)
	}

	return messages, nil
}
