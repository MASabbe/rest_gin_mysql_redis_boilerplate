package repository

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

// PermissionRepository defines persistence contracts for Permission entities.
type PermissionRepository interface {
	Create(ctx context.Context, perm *entity.Permission) error
	FindByID(ctx context.Context, id string) (*entity.Permission, error)
	FindByName(ctx context.Context, name string) (*entity.Permission, error)
	List(ctx context.Context, p pagination.Pagination) ([]*entity.Permission, int64, error)
	Update(ctx context.Context, perm *entity.Permission) error
	Delete(ctx context.Context, id string) error

	// Multi-lookup queries
	FindByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
	FindEffectivePermissionsByUserID(ctx context.Context, userID string) ([]string, error)
	GetAffectedUserIDsByPermissionID(ctx context.Context, permissionID string) ([]string, error)
}
