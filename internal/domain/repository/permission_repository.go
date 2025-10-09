package repository

import (
	"context"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	// Create creates a new permission in the repository
	Create(ctx context.Context, permission *entity.Permission) error

	// FindByID retrieves a permission by ID
	FindByID(ctx context.Context, id string) (*entity.Permission, error)

	// FindByName retrieves a permission by name
	FindByName(ctx context.Context, name string) (*entity.Permission, error)

	// List retrieves all permissions
	List(ctx context.Context) ([]*entity.Permission, error)
}
