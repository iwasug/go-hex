package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang-chi-hexagonal-base/internal/domain/entity"
)

// PermissionRepository implements the repository.PermissionRepository interface using PostgreSQL
type PermissionRepository struct {
	db *DB
}

// NewPermissionRepository creates a new PermissionRepository instance
func NewPermissionRepository(db *DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create creates a new permission in the database
func (r *PermissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	query := `
		INSERT INTO permissions (id, name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		permission.ID,
		permission.Name,
		permission.Resource,
		permission.Action,
		permission.Description,
		permission.CreatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("permission name already exists: %w", err)
		}
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// FindByID retrieves a permission by ID
func (r *PermissionRepository) FindByID(ctx context.Context, id string) (*entity.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE id = $1
	`

	permission := &entity.Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Resource,
		&permission.Action,
		&permission.Description,
		&permission.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find permission by ID: %w", err)
	}

	return permission, nil
}

// FindByName retrieves a permission by name
func (r *PermissionRepository) FindByName(ctx context.Context, name string) (*entity.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE name = $1
	`

	permission := &entity.Permission{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Resource,
		&permission.Action,
		&permission.Description,
		&permission.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find permission by name: %w", err)
	}

	return permission, nil
}

// List retrieves all permissions
func (r *PermissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*entity.Permission
	for rows.Next() {
		permission := &entity.Permission{}
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.Resource,
			&permission.Action,
			&permission.Description,
			&permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}
