package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

// RoleRepository implements the repository.RoleRepository interface using PostgreSQL
type RoleRepository struct {
	db *DB
}

// NewRoleRepository creates a new RoleRepository instance
func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role in the database
func (r *RoleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (id, name, description, is_deleted, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		role.ID,
		role.Name,
		role.Description,
		role.IsDeleted,
		role.CreatedBy,
		role.CreatedAt,
		role.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role name already exists: %w", err)
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// FindByID retrieves a role by ID (excludes soft deleted by default)
func (r *RoleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_deleted, created_by, updated_by, deleted_by,
		       created_at, updated_at, deleted_at
		FROM roles
		WHERE id = $1 AND is_deleted = false
	`

	role := &entity.Role{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsDeleted,
		&role.CreatedBy,
		&role.UpdatedBy,
		&role.DeletedBy,
		&role.CreatedAt,
		&role.UpdatedAt,
		&role.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find role by ID: %w", err)
	}

	return role, nil
}

// FindByName retrieves a role by name (excludes soft deleted by default)
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_deleted, created_by, updated_by, deleted_by,
		       created_at, updated_at, deleted_at
		FROM roles
		WHERE name = $1 AND is_deleted = false
	`

	role := &entity.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsDeleted,
		&role.CreatedBy,
		&role.UpdatedBy,
		&role.DeletedBy,
		&role.CreatedAt,
		&role.UpdatedAt,
		&role.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find role by name: %w", err)
	}

	return role, nil
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles
		SET name = $1, description = $2, updated_by = $3, updated_at = $4
		WHERE id = $5 AND is_deleted = false
	`

	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.Description,
		role.UpdatedBy,
		role.UpdatedAt,
		role.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role name already exists: %w", err)
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found or already deleted")
	}

	return nil
}

// Delete performs a soft delete on a role
func (r *RoleRepository) Delete(ctx context.Context, id string, deletedBy string) error {
	query := `
		UPDATE roles
		SET is_deleted = true, deleted_by = $1, deleted_at = $2, updated_at = $3
		WHERE id = $4 AND is_deleted = false
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, deletedBy, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found or already deleted")
	}

	return nil
}

// Restore restores a soft deleted role
func (r *RoleRepository) Restore(ctx context.Context, id string) error {
	query := `
		UPDATE roles
		SET is_deleted = false, deleted_by = NULL, deleted_at = NULL, updated_at = $1
		WHERE id = $2 AND is_deleted = true
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found or not deleted")
	}

	return nil
}

// List retrieves a list of roles with filtering and sorting
func (r *RoleRepository) List(ctx context.Context, params repository.ListRolesParams) ([]*entity.Role, error) {
	// Validate order_by field (whitelist)
	validOrderByFields := map[string]bool{
		"name":       true,
		"created_at": true,
	}

	orderBy := "name"
	if params.OrderBy != "" {
		if !validOrderByFields[params.OrderBy] {
			return nil, fmt.Errorf("invalid order_by field: %s", params.OrderBy)
		}
		orderBy = params.OrderBy
	}

	// Validate order direction
	order := "asc"
	if params.Order != "" {
		if params.Order != "asc" && params.Order != "desc" {
			return nil, fmt.Errorf("invalid order direction: %s", params.Order)
		}
		order = params.Order
	}

	// Build query with dynamic search and sorting
	query := `
		SELECT 
			r.id, r.name, r.description, r.is_deleted, 
			r.created_by, r.updated_by, r.deleted_by,
			r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		WHERE 
			($1 = '' OR r.name ILIKE '%' || $1 || '%' OR r.description ILIKE '%' || $1 || '%')
			AND ($2 = true OR r.is_deleted = false)
	`

	// Add ORDER BY clause using CASE statements for safety
	query += fmt.Sprintf(`
		ORDER BY 
			CASE WHEN '%s' = 'name' AND '%s' = 'asc' THEN r.name END ASC,
			CASE WHEN '%s' = 'name' AND '%s' = 'desc' THEN r.name END DESC,
			CASE WHEN '%s' = 'created_at' AND '%s' = 'asc' THEN r.created_at END ASC,
			CASE WHEN '%s' = 'created_at' AND '%s' = 'desc' THEN r.created_at END DESC
	`, orderBy, order, orderBy, order, orderBy, order, orderBy, order)

	search := strings.TrimSpace(params.Search)
	rows, err := r.db.QueryContext(ctx, query, search, params.IncludeDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		role := &entity.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsDeleted,
			&role.CreatedBy,
			&role.UpdatedBy,
			&role.DeletedBy,
			&role.CreatedAt,
			&role.UpdatedAt,
			&role.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
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

// AssignPermission assigns a permission to a role
func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, assigned_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("permission already assigned to role: %w", err)
		}
		if strings.Contains(err.Error(), "foreign key") {
			return fmt.Errorf("role or permission not found: %w", err)
		}
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	return nil
}

// RemovePermission removes a permission from a role
func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission assignment not found")
	}

	return nil
}

// IsRoleInUse checks if a role is currently assigned to any users
func (r *RoleRepository) IsRoleInUse(ctx context.Context, roleID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_roles WHERE role_id = $1
		)
	`

	var inUse bool
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(&inUse)
	if err != nil {
		return false, fmt.Errorf("failed to check if role is in use: %w", err)
	}

	return inUse, nil
}
