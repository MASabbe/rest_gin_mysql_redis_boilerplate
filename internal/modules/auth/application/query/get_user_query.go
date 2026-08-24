package query

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type GetUserByIDQuery struct {
	UserID string
}

func (q GetUserByIDQuery) Validate() error {
	if strings.TrimSpace(q.UserID) == "" {
		return appErrors.NewValidationError("user_id is required", appErrors.FieldError{Field: "user_id", Message: "cannot be empty"})
	}
	return nil
}
