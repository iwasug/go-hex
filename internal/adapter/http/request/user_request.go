package request

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AssignRoleRequest represents the request body for assigning a role to a user
type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}
