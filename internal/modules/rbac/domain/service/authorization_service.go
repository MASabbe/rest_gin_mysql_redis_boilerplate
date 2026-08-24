package service

import "context"

// AuthorizationService defines domain/application capability to evaluate user permissions.
type AuthorizationService interface {
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}
