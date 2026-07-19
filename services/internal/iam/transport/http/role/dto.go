package role

type createRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type updateRoleRequest struct {
	Description *string `json:"description"`
}

type assignRoleRequest struct {
	Role string `json:"role"`
}

type assignPermissionRequest struct {
	Permission string `json:"permission"`
}
