package repository

import (
	"context"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
)

// UserRepository defines the persistence contract for User entities.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error
	UpdateLastActivityAt(ctx context.Context, userID string, at time.Time) error
}
