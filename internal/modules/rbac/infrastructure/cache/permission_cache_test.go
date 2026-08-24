package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

func TestRedisPermissionCache_NilClientGracefulFallback(t *testing.T) {
	c := cache.NewRedisPermissionCache(nil)
	ctx := context.Background()

	perms, hit, err := c.Get(ctx, "user-1")
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, perms)

	err = c.Set(ctx, "user-1", []string{"user:read"}, time.Hour)
	assert.NoError(t, err)

	err = c.InvalidateUser(ctx, "user-1")
	assert.NoError(t, err)

	err = c.InvalidateUsers(ctx, []string{"user-1", "user-2"})
	assert.NoError(t, err)
}
