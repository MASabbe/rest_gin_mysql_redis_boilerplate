package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type LoginCommand struct {
	Email    string
	Password string
}

func (c LoginCommand) Validate() error {
	var errs []appErrors.FieldError
	if strings.TrimSpace(c.Email) == "" {
		errs = append(errs, appErrors.FieldError{Field: "email", Message: "email is required"})
	}
	if strings.TrimSpace(c.Password) == "" {
		errs = append(errs, appErrors.FieldError{Field: "password", Message: "password is required"})
	}
	if len(errs) > 0 {
		return appErrors.NewValidationError("login validation failed", errs...)
	}
	return nil
}
