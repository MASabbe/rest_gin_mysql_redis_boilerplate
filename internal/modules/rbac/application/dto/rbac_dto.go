package dto

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
)

type RoleDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToRoleDTO(r *entity.Role) RoleDTO {
	if r == nil {
		return RoleDTO{}
	}
	return RoleDTO{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func ToRoleDTOList(roles []*entity.Role) []RoleDTO {
	list := make([]RoleDTO, 0, len(roles))
	for _, r := range roles {
		list = append(list, ToRoleDTO(r))
	}
	return list
}

type PermissionDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToPermissionDTO(p *entity.Permission) PermissionDTO {
	if p == nil {
		return PermissionDTO{}
	}
	return PermissionDTO{
		ID:          p.ID,
		Name:        p.Name,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToPermissionDTOList(perms []*entity.Permission) []PermissionDTO {
	list := make([]PermissionDTO, 0, len(perms))
	for _, p := range perms {
		list = append(list, ToPermissionDTO(p))
	}
	return list
}

type RoleWithPermissionsDTO struct {
	Role        RoleDTO         `json:"role"`
	Permissions []PermissionDTO `json:"permissions"`
}

type UserRolesDTO struct {
	UserID string    `json:"user_id"`
	Roles  []RoleDTO `json:"roles"`
}

type UserPermissionsDTO struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
}
