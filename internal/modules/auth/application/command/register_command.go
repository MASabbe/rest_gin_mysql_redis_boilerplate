package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type RegisterCommand struct {
	Email    string
	Password string
}

func (c RegisterCommand) Validate() error {
	var errs []appErrors.FieldError
	if strings.TrimSpace(c.Email) == "" {
		errs = append(errs, appErrors.FieldError{Field: "email", Message: "email is required"})
	}
	if len(c.Password) < 8 {
		errs = append(errs, appErrors.FieldError{Field: "password", Message: "password must be at least 8 characters long"})
	}
	if len(errs) > 0 {
		return appErrors.NewValidationError("registration validation failed", errs...)
	}
	return nil
}
