package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type CreatePermissionCommand struct {
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

func (c CreatePermissionCommand) Validate() error {
	if strings.TrimSpace(c.Resource) == "" {
		return appErrors.NewValidationError("permission resource is required", appErrors.FieldError{
			Field:   "resource",
			Message: "cannot be empty",
		})
	}
	if strings.TrimSpace(c.Action) == "" {
		return appErrors.NewValidationError("permission action is required", appErrors.FieldError{
			Field:   "action",
			Message: "cannot be empty",
		})
	}
	return nil
}

type UpdatePermissionCommand struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

func (c UpdatePermissionCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("permission ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	if strings.TrimSpace(c.Resource) == "" {
		return appErrors.NewValidationError("permission resource is required", appErrors.FieldError{
			Field:   "resource",
			Message: "cannot be empty",
		})
	}
	if strings.TrimSpace(c.Action) == "" {
		return appErrors.NewValidationError("permission action is required", appErrors.FieldError{
			Field:   "action",
			Message: "cannot be empty",
		})
	}
	return nil
}

type DeletePermissionCommand struct {
	ID string `json:"id"`
}

func (c DeletePermissionCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("permission ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}
