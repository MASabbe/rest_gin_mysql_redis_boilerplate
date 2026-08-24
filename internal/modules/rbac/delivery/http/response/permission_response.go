package response

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/dto"
)

type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromPermissionDTO(dto *dto.PermissionDTO) PermissionResponse {
	if dto == nil {
		return PermissionResponse{}
	}
	return PermissionResponse{
		ID:          dto.ID,
		Name:        dto.Name,
		Resource:    dto.Resource,
		Action:      dto.Action,
		Description: dto.Description,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func FromPermissionDTOList(dtos []dto.PermissionDTO) []PermissionResponse {
	list := make([]PermissionResponse, 0, len(dtos))
	for _, d := range dtos {
		list = append(list, FromPermissionDTO(&d))
	}
	return list
}
