package response

import "time"

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	IsDeleted   bool                 `json:"is_deleted"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	DeletedAt   *time.Time           `json:"deleted_at,omitempty"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
