package response

import "time"

// UserResponse represents a user in API responses (without sensitive data)
type UserResponse struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Roles     []RoleResponse `json:"roles,omitempty"`
	IsDeleted bool           `json:"is_deleted"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty"`
}
