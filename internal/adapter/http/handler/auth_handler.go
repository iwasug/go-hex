package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang-chi-hexagonal-base/internal/adapter/http/request"
	"golang-chi-hexagonal-base/internal/adapter/http/response"
	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/service"
	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateEmail("email", req.Email)
	v.ValidatePassword("password", req.Password)
	v.ValidateRequired("name", req.Name)
	v.ValidateMaxLength("name", req.Name, 255)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	user, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})

	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			respondError(w, http.StatusConflict, "CONFLICT", "Email already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusCreated, toUserResponse(user))
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateEmail("email", req.Email)
	v.ValidateRequired("password", req.Password)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	loginResp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials", nil)
			return
		}
		if errors.Is(err, service.ErrUserDeleted) {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User account has been deleted", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to login", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, response.AuthResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		User:         toUserResponse(loginResp.User),
	})
}

// RefreshToken handles access token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req request.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("refresh_token", req.RefreshToken)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	accessToken, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid refresh token", nil)
			return
		}
		if errors.Is(err, service.ErrTokenExpired) {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Refresh token has expired", nil)
			return
		}
		if errors.Is(err, service.ErrUserDeleted) {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User account has been deleted", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to refresh token", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, response.RefreshTokenResponse{
		AccessToken: accessToken,
	})
}

// ForgotPassword handles forgot password request
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req request.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateEmail("email", req.Email)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	err := h.authService.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process forgot password request", nil)
		return
	}

	// Always return success for security reasons
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req request.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("token", req.Token)
	v.ValidatePassword("new_password", req.NewPassword)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid or expired reset token", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reset password", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "Password has been reset successfully",
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req request.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("refresh_token", req.RefreshToken)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	err := h.authService.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to logout", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "Logged out successfully",
	})
}

// toUserResponse converts entity.User to response.UserResponse
func toUserResponse(user *entity.User) response.UserResponse {
	return response.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsDeleted: user.IsDeleted,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
	}
}
