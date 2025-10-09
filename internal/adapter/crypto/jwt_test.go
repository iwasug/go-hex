package crypto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	token, err := jwtManager.GenerateAccessToken(userID, email, roles)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, AccessToken, claims.Type)
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	userID := "user-123"

	token, err := jwtManager.GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := jwtManager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, RefreshToken, claims.Type)
	assert.Empty(t, claims.Email)
	assert.Empty(t, claims.Roles)
}

func TestJWTManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Create manager with very short expiry
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 1*time.Millisecond, 1*time.Millisecond)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user"}

	token, err := jwtManager.GenerateAccessToken(userID, email, roles)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate should fail
	_, err = jwtManager.ValidateToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	// Test with invalid token string
	_, err := jwtManager.ValidateToken("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	jwtManager1 := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)
	jwtManager2 := NewJWTManager("different-secret-key-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"user"}

	// Generate token with first manager
	token, err := jwtManager1.GenerateAccessToken(userID, email, roles)
	require.NoError(t, err)

	// Try to validate with second manager (different secret)
	_, err = jwtManager2.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTManager_ParseToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"admin", "user"}

	token, err := jwtManager.GenerateAccessToken(userID, email, roles)
	require.NoError(t, err)

	// Parse token
	claims, err := jwtManager.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, AccessToken, claims.Type)
}

func TestJWTManager_TokenExpiry(t *testing.T) {
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", accessExpiry, refreshExpiry)

	// Test access token expiry
	accessToken, err := jwtManager.GenerateAccessToken("user-123", "test@example.com", []string{"user"})
	require.NoError(t, err)

	accessClaims, err := jwtManager.ValidateToken(accessToken)
	require.NoError(t, err)

	expectedAccessExpiry := time.Now().Add(accessExpiry)
	actualAccessExpiry := accessClaims.ExpiresAt.Time

	// Allow 1 second tolerance for test execution time
	assert.WithinDuration(t, expectedAccessExpiry, actualAccessExpiry, 1*time.Second)

	// Test refresh token expiry
	refreshToken, err := jwtManager.GenerateRefreshToken("user-123")
	require.NoError(t, err)

	refreshClaims, err := jwtManager.ValidateToken(refreshToken)
	require.NoError(t, err)

	expectedRefreshExpiry := time.Now().Add(refreshExpiry)
	actualRefreshExpiry := refreshClaims.ExpiresAt.Time

	// Allow 1 second tolerance for test execution time
	assert.WithinDuration(t, expectedRefreshExpiry, actualRefreshExpiry, 1*time.Second)
}

func TestJWTManager_TokenPayload(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 15*time.Minute, 7*24*time.Hour)

	t.Run("Access token includes user_id, email, and roles", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		roles := []string{"admin", "user", "moderator"}

		token, err := jwtManager.GenerateAccessToken(userID, email, roles)
		require.NoError(t, err)

		claims, err := jwtManager.ValidateToken(token)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roles, claims.Roles)
		assert.Len(t, claims.Roles, 3)
	})

	t.Run("Refresh token includes only user_id", func(t *testing.T) {
		userID := "user-456"

		token, err := jwtManager.GenerateRefreshToken(userID)
		require.NoError(t, err)

		claims, err := jwtManager.ValidateToken(token)
		require.NoError(t, err)

		assert.Equal(t, userID, claims.UserID)
		assert.Empty(t, claims.Email)
		assert.Empty(t, claims.Roles)
	})
}
