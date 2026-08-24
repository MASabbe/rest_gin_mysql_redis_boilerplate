package entity

import (
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/google/uuid"
)

// Role represents a user role in the system.
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewRole creates and validates a standard user-manageable role.
func NewRole(name, description string) (*Role, error) {
	return createRole(name, description, false)
}

// NewSystemRole creates and validates a system-protected role.
func NewSystemRole(name, description string) (*Role, error) {
	return createRole(name, description, true)
}

func createRole(name, description string, isSystem bool) (*Role, error) {
	cleanName := strings.ToLower(strings.TrimSpace(name))
	if cleanName == "" {
		return nil, appErrors.NewValidationError("role name is required", appErrors.FieldError{
			Field:   "name",
			Message: "role name cannot be empty",
		})
	}

	if len(cleanName) < 2 || len(cleanName) > 64 {
		return nil, appErrors.NewValidationError("role name must be between 2 and 64 characters", appErrors.FieldError{
			Field:   "name",
			Message: "length must be between 2 and 64 characters",
		})
	}

	now := time.Now().UTC()
	return &Role{
		ID:          uuid.New().String(),
		Name:        cleanName,
		Description: strings.TrimSpace(description),
		IsSystem:    isSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update updates the mutable fields of a role.
func (r *Role) Update(name, description string) error {
	cleanName := strings.ToLower(strings.TrimSpace(name))
	if cleanName == "" {
		return appErrors.NewValidationError("role name is required", appErrors.FieldError{
			Field:   "name",
			Message: "role name cannot be empty",
		})
	}

	if len(cleanName) < 2 || len(cleanName) > 64 {
		return appErrors.NewValidationError("role name must be between 2 and 64 characters", appErrors.FieldError{
			Field:   "name",
			Message: "length must be between 2 and 64 characters",
		})
	}

	r.Name = cleanName
	r.Description = strings.TrimSpace(description)
	r.UpdatedAt = time.Now().UTC()
	return nil
}
