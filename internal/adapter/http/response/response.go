package response

import "time"

// UserResponse represents a user in API responses (without sensitive data)
type UserResponse struct {
	ID        string          `json:"id"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Roles     []RoleResponse  `json:"roles,omitempty"`
	IsDeleted bool            `json:"is_deleted"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

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

// AuthResponse represents the response for login/register
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RefreshTokenResponse represents the response for refresh token
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}
