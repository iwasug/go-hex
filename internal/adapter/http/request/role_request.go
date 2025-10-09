package request

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRolePermissionsRequest represents the request body for updating role permissions
type UpdateRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}
