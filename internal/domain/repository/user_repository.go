package repository

import (
	"context"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// ListUsersParams contains parameters for listing users with filtering and pagination
type ListUsersParams struct {
	Search         string // search by name or email
	OrderBy        string // field to order by (e.g., "name", "email", "created_at")
	Order          string // "asc" or "desc"
	Limit          int
	Offset         int
	IncludeDeleted bool // include soft deleted records
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user in the repository
	Create(ctx context.Context, user *entity.User) error

	// FindByID retrieves a user by ID (excludes soft deleted by default)
	FindByID(ctx context.Context, id string) (*entity.User, error)

	// FindByEmail retrieves a user by email (excludes soft deleted by default)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entity.User) error

	// Delete performs a soft delete on a user
	Delete(ctx context.Context, id string, deletedBy string) error

	// Restore restores a soft deleted user
	Restore(ctx context.Context, id string) error

	// List retrieves a list of users with filtering, sorting, and pagination
	List(ctx context.Context, params ListUsersParams) ([]*entity.User, int, error)

	// GetUserRoles retrieves all roles assigned to a user
	GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error)

	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, userID, roleID string) error

	// RemoveRole removes a role from a user
	RemoveRole(ctx context.Context, userID, roleID string) error

	// GetUserPermissions retrieves all permissions for a user (through their roles)
	GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error)
}
