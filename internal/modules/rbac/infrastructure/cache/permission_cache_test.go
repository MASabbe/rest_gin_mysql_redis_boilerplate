package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/cache"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisPermissionCache_FullLifecycle_WithMiniRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	c := cache.NewRedisPermissionCache(rdb)
	ctx := context.Background()
	userID := "user-lifecycle-1"

	// 1. Cache Miss
	perms, hit, err := c.Get(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, perms)

	// 2. Cache Set
	expectedPerms := []string{"article:read", "article:create"}
	err = c.Set(ctx, userID, expectedPerms, 10*time.Minute)
	assert.NoError(t, err)

	// 3. Cache Hit
	perms, hit, err = c.Get(ctx, userID)
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.ElementsMatch(t, expectedPerms, perms)

	// 4. Invalidate Single User
	err = c.InvalidateUser(ctx, userID)
	assert.NoError(t, err)

	perms, hit, err = c.Get(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, perms)

	// 5. Invalidate Multiple Users
	_ = c.Set(ctx, "user-a", []string{"perm:a"}, time.Minute)
	_ = c.Set(ctx, "user-b", []string{"perm:b"}, time.Minute)

	err = c.InvalidateUsers(ctx, []string{"user-a", "user-b"})
	assert.NoError(t, err)

	_, hitA, _ := c.Get(ctx, "user-a")
	_, hitB, _ := c.Get(ctx, "user-b")
	assert.False(t, hitA)
	assert.False(t, hitB)

	// 6. TTL Expiration via MiniRedis Fast-Forward
	_ = c.Set(ctx, "user-ttl", []string{"perm:ttl"}, 5*time.Second)
	mr.FastForward(6 * time.Second)

	_, hitTTL, err := c.Get(ctx, "user-ttl")
	assert.NoError(t, err)
	assert.False(t, hitTTL)
}

func TestRedisPermissionCache_NilAndDegradedFallback(t *testing.T) {
	// Nil client
	cNil := cache.NewRedisPermissionCache(nil)
	ctx := context.Background()

	perms, hit, err := cNil.Get(ctx, "user-nil")
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, perms)

	assert.NoError(t, cNil.Set(ctx, "user-nil", []string{"perm"}, time.Hour))
	assert.NoError(t, cNil.InvalidateUser(ctx, "user-nil"))
	assert.NoError(t, cNil.InvalidateUsers(ctx, []string{"user-nil"}))

	// Degraded / closed client
	closedRdb := goredis.NewClient(&goredis.Options{
		Addr: "127.0.0.1:9999", // non-existent
	})
	cDegraded := cache.NewRedisPermissionCache(closedRdb)

	perms, hit, err = cDegraded.Get(ctx, "user-degraded")
	assert.Error(t, err) // Returns error on network failure to signal fallback to caller
	assert.False(t, hit)
	assert.Nil(t, perms)
}
