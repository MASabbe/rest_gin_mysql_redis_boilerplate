package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type CreateRoleCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c CreateRoleCommand) Validate() error {
	cleanName := strings.TrimSpace(c.Name)
	if cleanName == "" {
		return appErrors.NewValidationError("role name is required", appErrors.FieldError{
			Field:   "name",
			Message: "cannot be empty",
		})
	}
	if len(cleanName) < 2 || len(cleanName) > 64 {
		return appErrors.NewValidationError("role name length must be between 2 and 64 characters", appErrors.FieldError{
			Field:   "name",
			Message: "must be 2-64 characters",
		})
	}
	return nil
}

type UpdateRoleCommand struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c UpdateRoleCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("role ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	cleanName := strings.TrimSpace(c.Name)
	if cleanName == "" {
		return appErrors.NewValidationError("role name is required", appErrors.FieldError{
			Field:   "name",
			Message: "cannot be empty",
		})
	}
	if len(cleanName) < 2 || len(cleanName) > 64 {
		return appErrors.NewValidationError("role name length must be between 2 and 64 characters", appErrors.FieldError{
			Field:   "name",
			Message: "must be 2-64 characters",
		})
	}
	return nil
}

type DeleteRoleCommand struct {
	ID string `json:"id"`
}

func (c DeleteRoleCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("role ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type AssignRolePermissionsCommand struct {
	RoleID        string   `json:"role_id"`
	PermissionIDs []string `json:"permission_ids"`
}

func (c AssignRolePermissionsCommand) Validate() error {
	if strings.TrimSpace(c.RoleID) == "" {
		return appErrors.NewValidationError("role ID is required", appErrors.FieldError{
			Field:   "role_id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type AssignUserRolesCommand struct {
	UserID  string   `json:"user_id"`
	RoleIDs []string `json:"role_ids"`
}

func (c AssignUserRolesCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return appErrors.NewValidationError("user ID is required", appErrors.FieldError{
			Field:   "user_id",
			Message: "cannot be empty",
		})
	}
	return nil
}
