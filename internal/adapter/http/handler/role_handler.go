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
// @Summary List roles
// @Description Get a list of roles with optional search, sorting, and filtering
// @Tags Roles
// @Accept json
// @Produce json
// @Param search query string false "Search by name or description"
// @Param order_by query string false "Sort field (name, created_at)"
// @Param order query string false "Sort direction (asc, desc)"
// @Param include_deleted query boolean false "Include deleted roles (admin only)"
// @Success 200 {array} response.RoleResponse "List of roles"
// @Failure 400 {object} response.ErrorResponse "Invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires roles:read permission"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /roles [get]
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
	respondSuccess(w, http.StatusOK, "Roles retrieved successfully", roleResponses)
}

// Create handles creating a new role
// @Summary Create role
// @Description Create a new role with specified permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body request.CreateRoleRequest true "Role creation data"
// @Success 201 {object} response.RoleResponse "Role created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires roles:write permission"
// @Failure 409 {object} response.ErrorResponse "Role with this name already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /roles [post]
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
	respondSuccess(w, http.StatusCreated, "Role created successfully", toRoleResponse(role))
}

// GetByID handles retrieving a role by ID
// @Summary Get role by ID
// @Description Retrieve detailed information about a specific role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.RoleResponse "Role details"
// @Failure 400 {object} response.ErrorResponse "Invalid role ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires roles:read permission"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [get]
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
	respondSuccess(w, http.StatusOK, "Role retrieved successfully", toRoleResponse(role))
}

// Update handles updating a role
// @Summary Update role
// @Description Update role information (name and description)
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body request.UpdateRoleRequest true "Role update data"
// @Success 200 {object} response.RoleResponse "Updated role"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires roles:write permission"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 409 {object} response.ErrorResponse "Role with this name already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [put]
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
	respondSuccess(w, http.StatusOK, "Role retrieved successfully", toRoleResponse(role))
}

// Delete handles soft deleting a role
// @Summary Delete role
// @Description Soft delete a role (can be restored by admin)
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.MessageResponse "Role deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid role ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires roles:delete permission"
// @Failure 404 {object} response.ErrorResponse "Role not found"
// @Failure 409 {object} response.ErrorResponse "Role is in use and cannot be deleted"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [delete]
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
	respondSuccess(w, http.StatusOK, "Role deleted successfully", nil)
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
