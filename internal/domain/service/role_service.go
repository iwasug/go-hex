package service

import (
	"context"
	"errors"
	"fmt"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

var (
	// ErrRoleInUse is returned when trying to delete a role that is assigned to users
	ErrRoleInUse = errors.New("role is in use and cannot be deleted")
	// ErrPermissionNotFound is returned when permission is not found
	ErrPermissionNotFound = errors.New("permission not found")
)

// RoleService handles role management operations
type RoleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewRoleService creates a new RoleService instance
func NewRoleService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateRoleInput represents the input for creating a role
type CreateRoleInput struct {
	Name          string
	Description   string
	PermissionIDs []string
}

// Create creates a new role with permissions
func (s *RoleService) Create(ctx context.Context, input CreateRoleInput, createdBy string) (*entity.Role, error) {
	// Check if role with same name already exists
	existingRole, err := s.roleRepo.FindByName(ctx, input.Name)
	if err == nil && existingRole != nil {
		return nil, errors.New("role with this name already exists")
	}

	// Create role entity
	role := entity.NewRole(input.Name, input.Description, createdBy)

	// Create role in database
	err = s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions if provided
	if len(input.PermissionIDs) > 0 {
		for _, permissionID := range input.PermissionIDs {
			// Verify permission exists
			_, err := s.permissionRepo.FindByID(ctx, permissionID)
			if err != nil {
				// Skip invalid permissions but continue
				continue
			}

			err = s.roleRepo.AssignPermission(ctx, role.ID, permissionID)
			if err != nil {
				// Log error but continue with other permissions
				continue
			}
		}
	}

	return role, nil
}

// GetByID retrieves a role by ID
func (s *RoleService) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// ListRolesInput represents the input for listing roles
type ListRolesInput struct {
	Search         string
	OrderBy        string
	Order          string
	IncludeDeleted bool
}

// List retrieves a list of roles with search and sorting
func (s *RoleService) List(ctx context.Context, input ListRolesInput) ([]*entity.Role, error) {
	// Set defaults
	if input.OrderBy == "" {
		input.OrderBy = "name"
	}
	if input.Order == "" {
		input.Order = "asc"
	}

	// Build repository params
	params := repository.ListRolesParams{
		Search:         input.Search,
		OrderBy:        input.OrderBy,
		Order:          input.Order,
		IncludeDeleted: input.IncludeDeleted,
	}

	// Get roles from repository
	roles, err := s.roleRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// UpdateRoleInput represents the input for updating a role
type UpdateRoleInput struct {
	Name        string
	Description string
}

// Update updates a role's information
func (s *RoleService) Update(ctx context.Context, id string, input UpdateRoleInput, updatedBy string) (*entity.Role, error) {
	// Get existing role
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	// Check if name is being changed and if it's already in use
	if input.Name != role.Name {
		existingRole, err := s.roleRepo.FindByName(ctx, input.Name)
		if err == nil && existingRole != nil && existingRole.ID != id {
			return nil, errors.New("role with this name already exists")
		}
	}

	// Update role entity
	role.Update(input.Name, input.Description, updatedBy)

	// Save to repository
	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// Delete soft deletes a role
func (s *RoleService) Delete(ctx context.Context, id string, deletedBy string) error {
	// Check if role exists
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return ErrRoleNotFound
	}

	// Check if already deleted
	if role.IsDeleted {
		return ErrRoleNotFound
	}

	// Check if role is in use
	inUse, err := s.roleRepo.IsRoleInUse(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if role is in use: %w", err)
	}

	if inUse {
		return ErrRoleInUse
	}

	// Soft delete role
	err = s.roleRepo.Delete(ctx, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// Restore restores a soft deleted role
func (s *RoleService) Restore(ctx context.Context, id string) error {
	// Restore role
	err := s.roleRepo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore role: %w", err)
	}

	return nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	// Check if role exists
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	// Get role permissions
	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// UpdatePermissions updates the permissions assigned to a role
func (s *RoleService) UpdatePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Check if role exists
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	// Get current permissions
	currentPermissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get current permissions: %w", err)
	}

	// Create map of current permission IDs
	currentPermissionMap := make(map[string]bool)
	for _, perm := range currentPermissions {
		currentPermissionMap[perm.ID] = true
	}

	// Create map of new permission IDs
	newPermissionMap := make(map[string]bool)
	for _, permID := range permissionIDs {
		newPermissionMap[permID] = true
	}

	// Remove permissions that are no longer needed
	for _, perm := range currentPermissions {
		if !newPermissionMap[perm.ID] {
			err := s.roleRepo.RemovePermission(ctx, roleID, perm.ID)
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}

	// Add new permissions
	for _, permID := range permissionIDs {
		if !currentPermissionMap[permID] {
			// Verify permission exists
			_, err := s.permissionRepo.FindByID(ctx, permID)
			if err != nil {
				// Skip invalid permissions
				continue
			}

			err = s.roleRepo.AssignPermission(ctx, roleID, permID)
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}

	return nil
}
