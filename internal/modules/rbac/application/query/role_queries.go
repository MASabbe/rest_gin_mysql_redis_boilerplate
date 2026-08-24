package query

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

type GetRoleByIDQuery struct {
	ID string `json:"id"`
}

func (q GetRoleByIDQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return appErrors.NewValidationError("role ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type ListRolesQuery struct {
	Pagination pagination.Pagination
}

type GetRolePermissionsQuery struct {
	RoleID string `json:"role_id"`
}

func (q GetRolePermissionsQuery) Validate() error {
	if strings.TrimSpace(q.RoleID) == "" {
		return appErrors.NewValidationError("role ID is required", appErrors.FieldError{
			Field:   "role_id",
			Message: "cannot be empty",
		})
	}
	return nil
}
