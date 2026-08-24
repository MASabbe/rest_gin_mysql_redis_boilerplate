package query

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type GetUserRolesQuery struct {
	UserID string `json:"user_id"`
}

func (q GetUserRolesQuery) Validate() error {
	if strings.TrimSpace(q.UserID) == "" {
		return appErrors.NewValidationError("user ID is required", appErrors.FieldError{
			Field:   "user_id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type GetUserPermissionsQuery struct {
	UserID string `json:"user_id"`
}

func (q GetUserPermissionsQuery) Validate() error {
	if strings.TrimSpace(q.UserID) == "" {
		return appErrors.NewValidationError("user ID is required", appErrors.FieldError{
			Field:   "user_id",
			Message: "cannot be empty",
		})
	}
	return nil
}
