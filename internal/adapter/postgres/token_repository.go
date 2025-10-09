package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TokenRepository implements the repository.TokenRepository interface using PostgreSQL
type TokenRepository struct {
	db *DB
}

// NewTokenRepository creates a new TokenRepository instance
func NewTokenRepository(db *DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// SaveRefreshToken saves a refresh token for a user
func (r *TokenRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// FindRefreshToken retrieves the user ID associated with a refresh token
func (r *TokenRepository) FindRefreshToken(ctx context.Context, token string) (string, error) {
	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > $2
	`

	var userID string
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("refresh token not found or expired: %w", err)
		}
		return "", fmt.Errorf("failed to find refresh token: %w", err)
	}

	return userID, nil
}

// DeleteRefreshToken deletes a specific refresh token
func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

// DeleteUserRefreshTokens deletes all refresh tokens for a user
func (r *TokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}

	return nil
}

// SavePasswordResetToken saves a password reset token for a user
func (r *TokenRepository) SavePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save password reset token: %w", err)
	}

	return nil
}

// FindPasswordResetToken retrieves the user ID associated with a password reset token
func (r *TokenRepository) FindPasswordResetToken(ctx context.Context, token string) (string, error) {
	query := `
		SELECT user_id
		FROM password_reset_tokens
		WHERE token = $1 AND expires_at > $2
	`

	var userID string
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("password reset token not found or expired: %w", err)
		}
		return "", fmt.Errorf("failed to find password reset token: %w", err)
	}

	return userID, nil
}

// DeletePasswordResetToken deletes a specific password reset token
func (r *TokenRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE token = $1
	`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("password reset token not found")
	}

	return nil
}
