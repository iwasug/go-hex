package repository

import (
	"context"
	"time"
)

// TokenRepository defines the interface for token data operations
type TokenRepository interface {
	// SaveRefreshToken saves a refresh token for a user
	SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error

	// FindRefreshToken retrieves the user ID associated with a refresh token
	FindRefreshToken(ctx context.Context, token string) (string, error)

	// DeleteRefreshToken deletes a specific refresh token
	DeleteRefreshToken(ctx context.Context, token string) error

	// DeleteUserRefreshTokens deletes all refresh tokens for a user
	DeleteUserRefreshTokens(ctx context.Context, userID string) error

	// SavePasswordResetToken saves a password reset token for a user
	SavePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error

	// FindPasswordResetToken retrieves the user ID associated with a password reset token
	FindPasswordResetToken(ctx context.Context, token string) (string, error)

	// DeletePasswordResetToken deletes a specific password reset token
	DeletePasswordResetToken(ctx context.Context, token string) error
}
