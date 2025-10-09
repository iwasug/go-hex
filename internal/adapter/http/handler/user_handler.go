package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"golang-chi-hexagonal-base/internal/adapter/http/request"
	"golang-chi-hexagonal-base/internal/adapter/http/response"
	"golang-chi-hexagonal-base/internal/domain/service"
	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List handles listing users with search, sorting, and pagination
// @Summary List users
// @Description Get a paginated list of users with optional search, sorting, and filtering
// @Tags Users
// @Accept json
// @Produce json
// @Param search query string false "Search by name or email"
// @Param order_by query string false "Sort field (name, email, created_at)"
// @Param order query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param include_deleted query boolean false "Include deleted users (admin only)"
// @Success 200 {object} response.PaginatedResponse "List of users"
// @Failure 400 {object} response.ErrorResponse "Invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:read permission"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	search := r.URL.Query().Get("search")
	orderBy := r.URL.Query().Get("order_by")
	order := r.URL.Query().Get("order")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	includeDeletedStr := r.URL.Query().Get("include_deleted")

	// Parse page and limit
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse include_deleted
	includeDeleted := false
	if includeDeletedStr == "true" {
		includeDeleted = true
	}

	// Validate order_by field
	v := validator.New()
	allowedOrderByFields := []string{"name", "email", "created_at"}
	v.ValidateOrderBy("order_by", orderBy, allowedOrderByFields)
	v.ValidateOrderDirection("order", order)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	result, err := h.userService.List(r.Context(), service.ListUsersInput{
		Search:         search,
		OrderBy:        orderBy,
		Order:          order,
		Page:           page,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users", nil)
		return
	}

	// Convert users to response format
	userResponses := make([]response.UserResponse, len(result.Users))
	for i, user := range result.Users {
		userResponses[i] = toUserResponse(user)
	}

	// Return paginated response
	respondJSON(w, http.StatusOK, response.NewSuccessPaginatedResponse(
		"Users retrieved successfully",
		userResponses,
		response.Pagination{
			Total:      result.Total,
			Page:       result.Page,
			Limit:      result.Limit,
			TotalPages: result.TotalPages,
		},
	))
}

// GetByID handles retrieving a user by ID
// @Summary Get user by ID
// @Description Retrieve detailed information about a specific user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.UserResponse "User details"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:read permission"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "User ID is required", nil)
		return
	}

	// Call service
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user", nil)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

// Update handles updating a user
// @Summary Update user
// @Description Update user information (name and email)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body request.UpdateUserRequest true "User update data"
// @Success 200 {object} response.UserResponse "Updated user"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:write permission"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "User ID is required", nil)
		return
	}

	var req request.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMaxLength("name", req.Name, 255)
	v.ValidateEmail("email", req.Email)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Get updated_by from context (set by auth middleware)
	updatedBy := r.Context().Value("user_id")
	if updatedBy == nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in context", nil)
		return
	}

	// Call service
	user, err := h.userService.Update(r.Context(), id, service.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	}, updatedBy.(string))

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			respondError(w, http.StatusConflict, "CONFLICT", "Email already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user", nil)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

// Delete handles soft deleting a user
// @Summary Delete user
// @Description Soft delete a user (can be restored by admin)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.MessageResponse "User deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:delete permission"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "User ID is required", nil)
		return
	}

	// Get deleted_by from context (set by auth middleware)
	deletedBy := r.Context().Value("user_id")
	if deletedBy == nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in context", nil)
		return
	}

	// Call service
	err := h.userService.Delete(r.Context(), id, deletedBy.(string))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete user", nil)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "User deleted successfully", nil)
}

// AssignRole handles assigning a role to a user
// @Summary Assign role to user
// @Description Assign a role to a user for permission management
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body request.AssignRoleRequest true "Role assignment data"
// @Success 200 {object} response.MessageResponse "Role assigned successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:write permission"
// @Failure 404 {object} response.ErrorResponse "User or role not found"
// @Failure 409 {object} response.ErrorResponse "Role already assigned"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id}/roles [post]
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Validate user ID
	if userID == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "User ID is required", nil)
		return
	}

	var req request.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("role_id", req.RoleID)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	err := h.userService.AssignRole(r.Context(), userID, req.RoleID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "Role not found", nil)
			return
		}
		if errors.Is(err, service.ErrRoleAlreadyAssigned) {
			respondError(w, http.StatusConflict, "CONFLICT", "Role already assigned to user", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign role", nil)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Role assigned successfully", nil)
}

// RemoveRole handles removing a role from a user
// @Summary Remove role from user
// @Description Remove an assigned role from a user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 200 {object} response.MessageResponse "Role removed successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID or role ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires users:write permission"
// @Failure 404 {object} response.ErrorResponse "User or role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id}/roles/{role_id} [delete]
func (h *UserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	roleID := chi.URLParam(r, "role_id")

	// Validate IDs
	if userID == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "User ID is required", nil)
		return
	}
	if roleID == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Role ID is required", nil)
		return
	}

	// Call service
	err := h.userService.RemoveRole(r.Context(), userID, roleID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "Role not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove role", nil)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Role removed successfully", nil)
}
