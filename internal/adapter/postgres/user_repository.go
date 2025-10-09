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

// UserRepository implements the repository.UserRepository interface using PostgreSQL
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password, name, is_deleted, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		user.IsDeleted,
		user.CreatedBy,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("email already exists: %w", err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID (excludes soft deleted by default)
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, is_deleted, created_by, updated_by, deleted_by,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND is_deleted = false
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.IsDeleted,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.DeletedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return user, nil
}

// FindByEmail retrieves a user by email (excludes soft deleted by default)
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, is_deleted, created_by, updated_by, deleted_by,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND is_deleted = false
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.IsDeleted,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.DeletedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, updated_by = $4, updated_at = $5
		WHERE id = $6 AND is_deleted = false
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.UpdatedBy,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("email already exists: %w", err)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// Delete performs a soft delete on a user
func (r *UserRepository) Delete(ctx context.Context, id string, deletedBy string) error {
	query := `
		UPDATE users
		SET is_deleted = true, deleted_by = $1, deleted_at = $2, updated_at = $3
		WHERE id = $4 AND is_deleted = false
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, deletedBy, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// Restore restores a soft deleted user
func (r *UserRepository) Restore(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET is_deleted = false, deleted_by = NULL, deleted_at = NULL, updated_at = $1
		WHERE id = $2 AND is_deleted = true
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or not deleted")
	}

	return nil
}

// List retrieves a list of users with filtering, sorting, and pagination
func (r *UserRepository) List(ctx context.Context, params repository.ListUsersParams) ([]*entity.User, int, error) {
	// Validate order_by field (whitelist)
	validOrderByFields := map[string]bool{
		"name":       true,
		"email":      true,
		"created_at": true,
	}

	orderBy := "created_at"
	if params.OrderBy != "" {
		if !validOrderByFields[params.OrderBy] {
			return nil, 0, fmt.Errorf("invalid order_by field: %s", params.OrderBy)
		}
		orderBy = params.OrderBy
	}

	// Validate order direction
	order := "desc"
	if params.Order != "" {
		if params.Order != "asc" && params.Order != "desc" {
			return nil, 0, fmt.Errorf("invalid order direction: %s", params.Order)
		}
		order = params.Order
	}

	// Set default pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query with dynamic search and sorting
	query := `
		SELECT 
			u.id, u.email, u.password, u.name, u.is_deleted, 
			u.created_by, u.updated_by, u.deleted_by,
			u.created_at, u.updated_at, u.deleted_at,
			COUNT(*) OVER() as total_count
		FROM users u
		WHERE 
			($1 = '' OR u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')
			AND ($2 = true OR u.is_deleted = false)
	`

	// Add ORDER BY clause using CASE statements for safety
	query += fmt.Sprintf(`
		ORDER BY 
			CASE WHEN '%s' = 'name' AND '%s' = 'asc' THEN u.name END ASC,
			CASE WHEN '%s' = 'name' AND '%s' = 'desc' THEN u.name END DESC,
			CASE WHEN '%s' = 'email' AND '%s' = 'asc' THEN u.email END ASC,
			CASE WHEN '%s' = 'email' AND '%s' = 'desc' THEN u.email END DESC,
			CASE WHEN '%s' = 'created_at' AND '%s' = 'asc' THEN u.created_at END ASC,
			CASE WHEN '%s' = 'created_at' AND '%s' = 'desc' THEN u.created_at END DESC
		LIMIT $3 OFFSET $4
	`, orderBy, order, orderBy, order, orderBy, order, orderBy, order, orderBy, order, orderBy, order)

	search := strings.TrimSpace(params.Search)
	rows, err := r.db.QueryContext(ctx, query, search, params.IncludeDeleted, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	var totalCount int

	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.IsDeleted,
			&user.CreatedBy,
			&user.UpdatedBy,
			&user.DeletedBy,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}

// GetUserRoles retrieves all roles assigned to a user
func (r *UserRepository) GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_deleted, 
		       r.created_by, r.updated_by, r.deleted_by,
		       r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_deleted = false
		ORDER BY r.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
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

// AssignRole assigns a role to a user
func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, userID, roleID, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role already assigned to user: %w", err)
		}
		if strings.Contains(err.Error(), "foreign key") {
			return fmt.Errorf("user or role not found: %w", err)
		}
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user
func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role assignment not found")
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user (through their roles)
func (r *UserRepository) GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.is_deleted = false
		ORDER BY p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
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
