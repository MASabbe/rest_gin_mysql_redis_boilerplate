package response

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/dto"
)

type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromRoleDTO(dto *dto.RoleDTO) RoleResponse {
	if dto == nil {
		return RoleResponse{}
	}
	return RoleResponse{
		ID:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		IsSystem:    dto.IsSystem,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func FromRoleDTOList(dtos []dto.RoleDTO) []RoleResponse {
	list := make([]RoleResponse, 0, len(dtos))
	for _, d := range dtos {
		list = append(list, FromRoleDTO(&d))
	}
	return list
}

type RoleWithPermissionsResponse struct {
	Role        RoleResponse         `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}

func FromRoleWithPermissionsDTO(dto *dto.RoleWithPermissionsDTO) RoleWithPermissionsResponse {
	if dto == nil {
		return RoleWithPermissionsResponse{}
	}
	return RoleWithPermissionsResponse{
		Role:        FromRoleDTO(&dto.Role),
		Permissions: FromPermissionDTOList(dto.Permissions),
	}
}

type UserRolesResponse struct {
	UserID string         `json:"user_id"`
	Roles  []RoleResponse `json:"roles"`
}

func FromUserRolesDTO(dto *dto.UserRolesDTO) UserRolesResponse {
	if dto == nil {
		return UserRolesResponse{}
	}
	return UserRolesResponse{
		UserID: dto.UserID,
		Roles:  FromRoleDTOList(dto.Roles),
	}
}

type UserPermissionsResponse struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
}

func FromUserPermissionsDTO(dto *dto.UserPermissionsDTO) UserPermissionsResponse {
	if dto == nil {
		return UserPermissionsResponse{}
	}
	return UserPermissionsResponse{
		UserID:      dto.UserID,
		Permissions: dto.Permissions,
	}
}
