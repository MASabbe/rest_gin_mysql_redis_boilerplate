package request

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=255"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=255"`
}

type AssignRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

type AssignUserRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}
