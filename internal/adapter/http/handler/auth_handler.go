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
// @Summary Register a new user
// @Description Create a new user account with email, password, and name
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "Registration request"
// @Success 201 {object} response.UserResponse "User created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/register [post]
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
	respondSuccess(w, http.StatusCreated, "User registered successfully", toUserResponse(user))
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with email and password, returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login credentials"
// @Success 200 {object} response.AuthResponse "Login successful"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials or user deleted"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/login [post]
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
	respondSuccess(w, http.StatusOK, "Login successful", response.AuthResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		User:         toUserResponse(loginResp.User),
	})
}

// RefreshToken handles access token refresh
// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.RefreshTokenResponse "New access token"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
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
	respondSuccess(w, http.StatusOK, "Token refreshed successfully", response.RefreshTokenResponse{
		AccessToken: accessToken,
	})
}

// ForgotPassword handles forgot password request
// @Summary Request password reset
// @Description Send password reset token to user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.ForgotPasswordRequest true "Email address"
// @Success 200 {object} response.MessageResponse "Password reset email sent (if email exists)"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/forgot-password [post]
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
	respondSuccess(w, http.StatusOK, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} response.MessageResponse "Password reset successful"
// @Failure 400 {object} response.ErrorResponse "Invalid or expired reset token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/reset-password [post]
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
	respondSuccess(w, http.StatusOK, "Password has been reset successfully", nil)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate refresh token and logout user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.LogoutRequest true "Refresh token"
// @Success 200 {object} response.MessageResponse "Logout successful"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
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
	respondSuccess(w, http.StatusOK, "Logged out successfully", nil)
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
