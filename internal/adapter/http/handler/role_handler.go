package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"golang-chi-hexagonal-base/internal/adapter/http/request"
	"golang-chi-hexagonal-base/internal/adapter/http/response"
	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/service"
	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new RoleHandler instance
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// List handles listing roles with search and sorting
// GET /api/v1/roles
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	search := r.URL.Query().Get("search")
	orderBy := r.URL.Query().Get("order_by")
	order := r.URL.Query().Get("order")
	includeDeletedStr := r.URL.Query().Get("include_deleted")

	// Parse include_deleted
	includeDeleted := false
	if includeDeletedStr == "true" {
		includeDeleted = true
	}

	// Validate order_by field
	v := validator.New()
	allowedOrderByFields := []string{"name", "created_at"}
	v.ValidateOrderBy("order_by", orderBy, allowedOrderByFields)
	v.ValidateOrderDirection("order", order)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	roles, err := h.roleService.List(r.Context(), service.ListRolesInput{
		Search:         search,
		OrderBy:        orderBy,
		Order:          order,
		IncludeDeleted: includeDeleted,
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list roles", nil)
		return
	}

	// Convert roles to response format
	roleResponses := make([]response.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = toRoleResponse(role)
	}

	// Return response
	respondJSON(w, http.StatusOK, roleResponses)
}

// Create handles creating a new role
// POST /api/v1/roles
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMaxLength("name", req.Name, 100)
	v.ValidateMaxLength("description", req.Description, 1000)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Get created_by from context (set by auth middleware)
	createdBy := r.Context().Value("user_id")
	if createdBy == nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in context", nil)
		return
	}

	// Call service
	role, err := h.roleService.Create(r.Context(), service.CreateRoleInput{
		Name:          req.Name,
		Description:   req.Description,
		PermissionIDs: req.PermissionIDs,
	}, createdBy.(string))

	if err != nil {
		if err.Error() == "role with this name already exists" {
			respondError(w, http.StatusConflict, "CONFLICT", "Role with this name already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create role", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusCreated, toRoleResponse(role))
}

// GetByID handles retrieving a role by ID
// GET /api/v1/roles/:id
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Role ID is required", nil)
		return
	}

	// Call service
	role, err := h.roleService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "Role not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get role", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, toRoleResponse(role))
}

// Update handles updating a role
// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Role ID is required", nil)
		return
	}

	var req request.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMaxLength("name", req.Name, 100)
	v.ValidateMaxLength("description", req.Description, 1000)

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
	role, err := h.roleService.Update(r.Context(), id, service.UpdateRoleInput{
		Name:        req.Name,
		Description: req.Description,
	}, updatedBy.(string))

	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "Role not found", nil)
			return
		}
		if err.Error() == "role with this name already exists" {
			respondError(w, http.StatusConflict, "CONFLICT", "Role with this name already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update role", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, toRoleResponse(role))
}

// Delete handles soft deleting a role
// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate ID
	if id == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Role ID is required", nil)
		return
	}

	// Get deleted_by from context (set by auth middleware)
	deletedBy := r.Context().Value("user_id")
	if deletedBy == nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in context", nil)
		return
	}

	// Call service
	err := h.roleService.Delete(r.Context(), id, deletedBy.(string))
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "Role not found", nil)
			return
		}
		if errors.Is(err, service.ErrRoleInUse) {
			respondError(w, http.StatusConflict, "CONFLICT", "Role is in use and cannot be deleted", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete role", nil)
		return
	}

	// Return response
	respondJSON(w, http.StatusOK, response.MessageResponse{
		Message: "Role deleted successfully",
	})
}

// toRoleResponse converts entity.Role to response.RoleResponse
func toRoleResponse(role *entity.Role) response.RoleResponse {
	return response.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsDeleted:   role.IsDeleted,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		DeletedAt:   role.DeletedAt,
	}
}
