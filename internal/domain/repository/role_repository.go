package repository

import (
	"context"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// ListRolesParams contains parameters for listing roles with filtering
type ListRolesParams struct {
	Search         string // search by name or description
	OrderBy        string // field to order by (e.g., "name", "created_at")
	Order          string // "asc" or "desc"
	IncludeDeleted bool   // include soft deleted records
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// Create creates a new role in the repository
	Create(ctx context.Context, role *entity.Role) error

	// FindByID retrieves a role by ID (excludes soft deleted by default)
	FindByID(ctx context.Context, id string) (*entity.Role, error)

	// FindByName retrieves a role by name (excludes soft deleted by default)
	FindByName(ctx context.Context, name string) (*entity.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *entity.Role) error

	// Delete performs a soft delete on a role
	Delete(ctx context.Context, id string, deletedBy string) error

	// Restore restores a soft deleted role
	Restore(ctx context.Context, id string) error

	// List retrieves a list of roles with filtering and sorting
	List(ctx context.Context, params ListRolesParams) ([]*entity.Role, error)

	// GetRolePermissions retrieves all permissions assigned to a role
	GetRolePermissions(ctx context.Context, roleID string) ([]*entity.Permission, error)

	// AssignPermission assigns a permission to a role
	AssignPermission(ctx context.Context, roleID, permissionID string) error

	// RemovePermission removes a permission from a role
	RemovePermission(ctx context.Context, roleID, permissionID string) error

	// IsRoleInUse checks if a role is currently assigned to any users
	IsRoleInUse(ctx context.Context, roleID string) (bool, error)
}
