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

// WhatsAppSessionRepository implements the repository.WhatsAppSessionRepository interface using PostgreSQL
type WhatsAppSessionRepository struct {
	db *DB
}

// NewWhatsAppSessionRepository creates a new WhatsAppSessionRepository instance
func NewWhatsAppSessionRepository(db *DB) *WhatsAppSessionRepository {
	return &WhatsAppSessionRepository{db: db}
}

// Create creates a new WhatsApp session in the database
func (r *WhatsAppSessionRepository) Create(ctx context.Context, session *entity.WhatsAppSession) error {
	query := `
		INSERT INTO whatsapp_sessions (
			id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
			device_info, last_activity, webhook_url, webhook_secret, is_active,
			created_by, updated_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.SessionName,
		session.PhoneNumber,
		session.Status,
		session.QRCode,
		session.QRCodeExpiry,
		session.DeviceInfo,
		session.LastActivity,
		session.WebhookURL,
		session.WebhookSecret,
		session.IsActive,
		session.CreatedBy,
		session.UpdatedBy,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("session with this name already exists for user: %w", err)
		}
		if strings.Contains(err.Error(), "foreign key") {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to create whatsapp session: %w", err)
	}

	return nil
}

// FindByID retrieves a WhatsApp session by ID
func (r *WhatsAppSessionRepository) FindByID(ctx context.Context, id string) (*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
		       device_info, last_activity, webhook_url, webhook_secret, is_active,
		       reconnect_attempts, last_reconnect_attempt,
		       created_by, updated_by, created_at, updated_at
		FROM whatsapp_sessions
		WHERE id = $1
	`

	session := &entity.WhatsAppSession{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionName,
		&session.PhoneNumber,
		&session.Status,
		&session.QRCode,
		&session.QRCodeExpiry,
		&session.DeviceInfo,
		&session.LastActivity,
		&session.WebhookURL,
		&session.WebhookSecret,
		&session.IsActive,
		&session.ReconnectAttempts,
		&session.LastReconnectAttempt,
		&session.CreatedBy,
		&session.UpdatedBy,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("whatsapp session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find whatsapp session by ID: %w", err)
	}

	return session, nil
}

// FindByUserID retrieves all WhatsApp sessions for a specific user
func (r *WhatsAppSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
		       device_info, last_activity, webhook_url, webhook_secret, is_active,
		       reconnect_attempts, last_reconnect_attempt,
		       created_by, updated_by, created_at, updated_at
		FROM whatsapp_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions by user ID: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// FindByUserIDAndName retrieves a session by user ID and session name
func (r *WhatsAppSessionRepository) FindByUserIDAndName(ctx context.Context, userID, sessionName string) (*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
		       device_info, last_activity, webhook_url, webhook_secret, is_active,
		       reconnect_attempts, last_reconnect_attempt,
		       created_by, updated_by, created_at, updated_at
		FROM whatsapp_sessions
		WHERE user_id = $1 AND session_name = $2
	`

	session := &entity.WhatsAppSession{}
	err := r.db.QueryRowContext(ctx, query, userID, sessionName).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionName,
		&session.PhoneNumber,
		&session.Status,
		&session.QRCode,
		&session.QRCodeExpiry,
		&session.DeviceInfo,
		&session.LastActivity,
		&session.WebhookURL,
		&session.WebhookSecret,
		&session.IsActive,
		&session.ReconnectAttempts,
		&session.LastReconnectAttempt,
		&session.CreatedBy,
		&session.UpdatedBy,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("whatsapp session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find whatsapp session by user ID and name: %w", err)
	}

	return session, nil
}

// Update updates an existing WhatsApp session
func (r *WhatsAppSessionRepository) Update(ctx context.Context, session *entity.WhatsAppSession) error {
	query := `
		UPDATE whatsapp_sessions
		SET session_name = $1, phone_number = $2, status = $3, qr_code = $4, qr_code_expiry = $5,
		    device_info = $6, last_activity = $7, webhook_url = $8, webhook_secret = $9,
		    is_active = $10, updated_by = $11, updated_at = $12
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		session.SessionName,
		session.PhoneNumber,
		session.Status,
		session.QRCode,
		session.QRCodeExpiry,
		session.DeviceInfo,
		session.LastActivity,
		session.WebhookURL,
		session.WebhookSecret,
		session.IsActive,
		session.UpdatedBy,
		session.UpdatedAt,
		session.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("session with this name already exists for user: %w", err)
		}
		return fmt.Errorf("failed to update whatsapp session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// Delete deletes a WhatsApp session from the repository
func (r *WhatsAppSessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM whatsapp_sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete whatsapp session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// List retrieves a list of WhatsApp sessions with filtering, sorting, and pagination
func (r *WhatsAppSessionRepository) List(ctx context.Context, params repository.ListSessionsParams) ([]*entity.WhatsAppSession, int, error) {
	// Validate order_by field (whitelist)
	validOrderByFields := map[string]bool{
		"session_name":  true,
		"phone_number":  true,
		"status":        true,
		"last_activity": true,
		"created_at":    true,
	}

	orderBy := "created_at"
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
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query with dynamic filtering
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT 
			id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
			device_info, last_activity, webhook_url, webhook_secret, is_active,
			reconnect_attempts, last_reconnect_attempt,
			created_by, updated_by, created_at, updated_at,
			COUNT(*) OVER() as total_count
		FROM whatsapp_sessions
		WHERE 1=1
	`)

	args := []interface{}{}
	argPos := 1

	// Filter by user ID (unless admin viewing all)
	if !params.IncludeAll && params.UserID != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND user_id = $%d", argPos))
		args = append(args, params.UserID)
		argPos++
	}

	// Filter by status
	if params.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argPos))
		args = append(args, params.Status)
		argPos++
	}

	// Filter by active status
	if params.IsActive != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND is_active = $%d", argPos))
		args = append(args, *params.IsActive)
		argPos++
	}

	// Search by session name or phone number
	if params.Search != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND (session_name ILIKE $%d OR phone_number ILIKE $%d)", argPos, argPos))
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
		return nil, 0, fmt.Errorf("failed to list whatsapp sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.WhatsAppSession
	var totalCount int

	for rows.Next() {
		session := &entity.WhatsAppSession{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SessionName,
			&session.PhoneNumber,
			&session.Status,
			&session.QRCode,
			&session.QRCodeExpiry,
			&session.DeviceInfo,
			&session.LastActivity,
			&session.WebhookURL,
			&session.WebhookSecret,
			&session.IsActive,
			&session.ReconnectAttempts,
			&session.LastReconnectAttempt,
			&session.CreatedBy,
			&session.UpdatedBy,
			&session.CreatedAt,
			&session.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan whatsapp session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating whatsapp sessions: %w", err)
	}

	return sessions, totalCount, nil
}

// UpdateStatus updates the status of a WhatsApp session
func (r *WhatsAppSessionRepository) UpdateStatus(ctx context.Context, id string, status entity.SessionStatus) error {
	query := `
		UPDATE whatsapp_sessions
		SET status = $1, last_activity = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// UpdateQRCode updates the QR code and expiry time for a session
func (r *WhatsAppSessionRepository) UpdateQRCode(ctx context.Context, id string, qrCode string, expiry time.Time) error {
	query := `
		UPDATE whatsapp_sessions
		SET qr_code = $1, qr_code_expiry = $2, status = $3, updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, qrCode, expiry, entity.SessionStatusAuthenticating, now, id)
	if err != nil {
		return fmt.Errorf("failed to update QR code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// ClearQRCode clears the QR code data for a session
func (r *WhatsAppSessionRepository) ClearQRCode(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_sessions
		SET qr_code = NULL, qr_code_expiry = NULL, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to clear QR code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// UpdatePhoneNumber updates the connected phone number for a session
func (r *WhatsAppSessionRepository) UpdatePhoneNumber(ctx context.Context, id string, phoneNumber string) error {
	query := `
		UPDATE whatsapp_sessions
		SET phone_number = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, phoneNumber, now, id)
	if err != nil {
		return fmt.Errorf("failed to update phone number: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// UpdateLastActivity updates the last activity timestamp for a session
func (r *WhatsAppSessionRepository) UpdateLastActivity(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_sessions
		SET last_activity = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// UpdateDeviceInfo updates the device information for a session
func (r *WhatsAppSessionRepository) UpdateDeviceInfo(ctx context.Context, id string, deviceInfo string) error {
	query := `
		UPDATE whatsapp_sessions
		SET device_info = $1, updated_at = $2
		WHERE id = $3
	`

	// Convert to pointer for nullable field
	var deviceInfoPtr *string
	if deviceInfo != "" {
		deviceInfoPtr = &deviceInfo
	}

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, deviceInfoPtr, now, id)
	if err != nil {
		return fmt.Errorf("failed to update device info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// FindActiveByUserID retrieves all active WhatsApp sessions for a specific user
func (r *WhatsAppSessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
		       device_info, last_activity, webhook_url, webhook_secret, is_active,
		       reconnect_attempts, last_reconnect_attempt,
		       created_by, updated_by, created_at, updated_at
		FROM whatsapp_sessions
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active sessions by user ID: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// FindConnectedSessions retrieves all sessions with connected status
func (r *WhatsAppSessionRepository) FindConnectedSessions(ctx context.Context) ([]*entity.WhatsAppSession, error) {
	query := `
		SELECT id, user_id, session_name, phone_number, status, qr_code, qr_code_expiry,
		       device_info, last_activity, webhook_url, webhook_secret, is_active,
		       reconnect_attempts, last_reconnect_attempt,
		       created_by, updated_by, created_at, updated_at
		FROM whatsapp_sessions
		WHERE status = $1 AND is_active = true
		ORDER BY last_activity DESC
	`

	rows, err := r.db.QueryContext(ctx, query, entity.SessionStatusConnected)
	if err != nil {
		return nil, fmt.Errorf("failed to find connected sessions: %w", err)
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// CountByUserID counts the number of sessions for a specific user
func (r *WhatsAppSessionRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM whatsapp_sessions WHERE user_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions by user ID: %w", err)
	}

	return count, nil
}

// IncrementReconnectAttempts increments the reconnection attempt counter
func (r *WhatsAppSessionRepository) IncrementReconnectAttempts(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_sessions
		SET reconnect_attempts = reconnect_attempts + 1,
		    last_reconnect_attempt = $1,
		    updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to increment reconnect attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// ResetReconnectAttempts resets the reconnection attempt counter
func (r *WhatsAppSessionRepository) ResetReconnectAttempts(ctx context.Context, id string) error {
	query := `
		UPDATE whatsapp_sessions
		SET reconnect_attempts = 0,
		    last_reconnect_attempt = NULL,
		    updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to reset reconnect attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("whatsapp session not found")
	}

	return nil
}

// scanSessions is a helper function to scan multiple sessions from rows
func (r *WhatsAppSessionRepository) scanSessions(rows *sql.Rows) ([]*entity.WhatsAppSession, error) {
	var sessions []*entity.WhatsAppSession

	for rows.Next() {
		session := &entity.WhatsAppSession{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SessionName,
			&session.PhoneNumber,
			&session.Status,
			&session.QRCode,
			&session.QRCodeExpiry,
			&session.DeviceInfo,
			&session.LastActivity,
			&session.WebhookURL,
			&session.WebhookSecret,
			&session.IsActive,
			&session.ReconnectAttempts,
			&session.LastReconnectAttempt,
			&session.CreatedBy,
			&session.UpdatedBy,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan whatsapp session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating whatsapp sessions: %w", err)
	}

	return sessions, nil
}
