package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang-chi-hexagonal-base/internal/adapter/crypto"
	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

var (
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserDeleted is returned when trying to login with a deleted user
	ErrUserDeleted = errors.New("user account has been deleted")
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")
	// ErrEmailAlreadyExists is returned when email is already registered
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	tokenRepo     repository.TokenRepository
	passwordHasher crypto.PasswordHasher
	tokenManager   crypto.TokenManager
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	tokenRepo repository.TokenRepository,
	passwordHasher crypto.PasswordHasher,
	tokenManager crypto.TokenManager,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		tokenRepo:     tokenRepo,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
	}
}

// RegisterInput represents the input for user registration
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// LoginResponse represents the response from login
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	User         *entity.User
}

// Register creates a new user account with default role
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*entity.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := entity.NewUser(input.Email, hashedPassword, input.Name)
	// created_by is nil for self-registration (already set by NewUser)

	// Create user in database
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default "user" role
	defaultRole, err := s.roleRepo.FindByName(ctx, "user")
	if err != nil {
		// If default role doesn't exist, log but don't fail registration
		// This allows the system to work even if roles aren't seeded yet
		return user, nil
	}

	err = s.userRepo.AssignRole(ctx, user.ID, defaultRole.ID)
	if err != nil {
		// Log error but don't fail registration
		// User can be assigned role later by admin
		return user, nil
	}

	return user, nil
}

// Login authenticates a user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is deleted
	if user.IsDeleted {
		return nil, ErrUserDeleted
	}

	// Verify password
	err = s.passwordHasher.ComparePassword(user.Password, password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Get user roles
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Convert roles to string slice
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate access token
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Save refresh token to database with 7 day expiry
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Find refresh token in database
	userID, err := s.tokenRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Validate refresh token
	claims, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		if errors.Is(err, crypto.ErrExpiredToken) {
			// Delete expired token
			_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
			return "", ErrTokenExpired
		}
		return "", ErrInvalidToken
	}

	// Verify token type
	if claims.Type != crypto.RefreshToken {
		return "", ErrInvalidToken
	}

	// Get user data
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is deleted
	if user.IsDeleted {
		return "", ErrUserDeleted
	}

	// Get user roles
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get user roles: %w", err)
	}

	// Convert roles to string slice
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate new access token
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, user.Email, roleNames)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// ForgotPassword generates a password reset token
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Return success even if user not found (security best practice)
		return nil
	}

	// Don't send reset token for deleted users
	if user.IsDeleted {
		return nil
	}

	// Generate reset token
	resetToken, err := generateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Save reset token with 1 hour expiry
	expiresAt := time.Now().Add(1 * time.Hour)
	err = s.tokenRepo.SavePasswordResetToken(ctx, user.ID, resetToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// TODO: Send email with reset token
	// In a real application, you would send an email here
	// For now, the token is just stored in the database

	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	// Find user by reset token
	userID, err := s.tokenRepo.FindPasswordResetToken(ctx, resetToken)
	if err != nil {
		return ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.UpdatePassword(hashedPassword, userID)
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	err = s.tokenRepo.DeletePasswordResetToken(ctx, resetToken)
	if err != nil {
		// Log error but don't fail the operation
		return nil
	}

	return nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
