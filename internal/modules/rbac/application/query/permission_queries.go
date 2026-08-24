package query

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

type GetPermissionByIDQuery struct {
	ID string `json:"id"`
}

func (q GetPermissionByIDQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return appErrors.NewValidationError("permission ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type ListPermissionsQuery struct {
	Pagination pagination.Pagination
}
