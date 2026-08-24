package repository

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

// RoleRepository defines persistence contracts for Role entities and user-role relations.
type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	FindByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context, p pagination.Pagination) ([]*entity.Role, int64, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error

	// Role-Permission associations
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissionIDs(ctx context.Context, roleID string) ([]string, error)

	// User-Role associations
	AssignUserRoles(ctx context.Context, userID string, roleIDs []string) error
	GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error)
	GetUserIDsByRoleID(ctx context.Context, roleID string) ([]string, error)
}
