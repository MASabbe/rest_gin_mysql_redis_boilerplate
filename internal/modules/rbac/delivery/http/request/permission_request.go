package request

type CreatePermissionRequest struct {
	Resource    string `json:"resource" binding:"required,min=2,max=64"`
	Action      string `json:"action" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=255"`
}

type UpdatePermissionRequest struct {
	Resource    string `json:"resource" binding:"required,min=2,max=64"`
	Action      string `json:"action" binding:"required,min=2,max=64"`
	Description string `json:"description" binding:"max=255"`
}
