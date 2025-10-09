package crypto

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken represents an access token
	AccessToken TokenType = "access"
	// RefreshToken represents a refresh token
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Roles  []string  `json:"roles"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenManager defines the interface for JWT token operations
type TokenManager interface {
	GenerateAccessToken(userID, email string, roles []string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ParseToken(tokenString string) (*Claims, error)
}

// JWTManager implements TokenManager using JWT
type JWTManager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTManager creates a new JWTManager
func NewJWTManager(secret string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// GenerateAccessToken generates a new access token with 15 minute expiry
func (j *JWTManager) GenerateAccessToken(userID, email string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		Type:   AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken generates a new refresh token with 7 day expiry
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates a token and returns the claims if valid
// Returns error if token is invalid or expired
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token has expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ParseToken parses a token and extracts the claims
// Does not validate expiry - use ValidateToken for full validation
func (j *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
