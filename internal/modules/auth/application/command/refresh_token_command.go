package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type RefreshTokenCommand struct {
	RefreshToken string
}

func (c RefreshTokenCommand) Validate() error {
	if strings.TrimSpace(c.RefreshToken) == "" {
		return appErrors.NewValidationError("refresh_token is required", appErrors.FieldError{Field: "refresh_token", Message: "cannot be empty"})
	}
	return nil
}
