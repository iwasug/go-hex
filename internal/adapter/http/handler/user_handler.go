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
// GET /api/v1/users
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
	respondJSON(w, http.StatusOK, response.PaginatedResponse{
		Data: userResponses,
		Pagination: response.Pagination{
			Total:      result.Total,
			Page:       result.Page,
			Limit:      result.Limit,
			TotalPages: result.TotalPages,
		},
	})
}

// GetByID handles retrieving a user by ID
// GET /api/v1/users/:id
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
	respondJSON(w, http.StatusOK, toUserResponse(user))
}

// Update handles updating a user
// PUT /api/v1/users/:id
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
	respondJSON(w, http.StatusOK, toUserResponse(user))
}

// Delete handles soft deleting a user
// DELETE /api/v1/users/:id
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
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "User deleted successfully",
	})
}

// AssignRole handles assigning a role to a user
// POST /api/v1/users/:id/roles
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
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "Role assigned successfully",
	})
}

// RemoveRole handles removing a role from a user
// DELETE /api/v1/users/:id/roles/:role_id
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
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "Role removed successfully",
	})
}
