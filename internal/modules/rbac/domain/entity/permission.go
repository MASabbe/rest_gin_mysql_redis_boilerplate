package entity

import (
	"fmt"
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/google/uuid"
)

// Permission represents an atomic authorization permission formatted as resource:action.
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"` // resource:action
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewPermission creates and validates a new Permission entity with deterministic resource:action naming.
func NewPermission(resource, action, description string) (*Permission, error) {
	cleanResource := strings.ToLower(strings.TrimSpace(resource))
	cleanAction := strings.ToLower(strings.TrimSpace(action))

	if cleanResource == "" {
		return nil, appErrors.NewValidationError("permission resource is required", appErrors.FieldError{
			Field:   "resource",
			Message: "resource cannot be empty",
		})
	}

	if cleanAction == "" {
		return nil, appErrors.NewValidationError("permission action is required", appErrors.FieldError{
			Field:   "action",
			Message: "action cannot be empty",
		})
	}

	deterministicName := fmt.Sprintf("%s:%s", cleanResource, cleanAction)
	now := time.Now().UTC()

	return &Permission{
		ID:          uuid.New().String(),
		Name:        deterministicName,
		Resource:    cleanResource,
		Action:      cleanAction,
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update updates the permission fields.
func (p *Permission) Update(resource, action, description string) error {
	cleanResource := strings.ToLower(strings.TrimSpace(resource))
	cleanAction := strings.ToLower(strings.TrimSpace(action))

	if cleanResource == "" {
		return appErrors.NewValidationError("permission resource is required", appErrors.FieldError{
			Field:   "resource",
			Message: "resource cannot be empty",
		})
	}

	if cleanAction == "" {
		return appErrors.NewValidationError("permission action is required", appErrors.FieldError{
			Field:   "action",
			Message: "action cannot be empty",
		})
	}

	p.Resource = cleanResource
	p.Action = cleanAction
	p.Name = fmt.Sprintf("%s:%s", cleanResource, cleanAction)
	p.Description = strings.TrimSpace(description)
	p.UpdatedAt = time.Now().UTC()
	return nil
}
