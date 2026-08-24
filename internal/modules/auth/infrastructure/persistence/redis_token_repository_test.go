package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisTokenRepository_Lifecycle_WithMiniRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	repo := persistence.NewRedisTokenRepository(rdb)
	ctx := context.Background()

	tokenID := "token-uuid-1"
	userID := "user-uuid-100"

	// 1. Store Refresh Token
	err = repo.StoreRefreshToken(ctx, tokenID, userID, time.Hour)
	assert.NoError(t, err)

	// 2. Validate Active Refresh Token
	foundUserID, err := repo.ValidateRefreshToken(ctx, tokenID)
	assert.NoError(t, err)
	assert.Equal(t, userID, foundUserID)

	// 3. Revoke Refresh Token
	err = repo.RevokeRefreshToken(ctx, tokenID)
	assert.NoError(t, err)

	// 4. Validate Revoked Refresh Token returns error
	_, err = repo.ValidateRefreshToken(ctx, tokenID)
	assert.Error(t, err)

	// 5. Revoke All User Tokens
	token1 := "user-token-1"
	token2 := "user-token-2"
	_ = repo.StoreRefreshToken(ctx, token1, userID, time.Hour)
	_ = repo.StoreRefreshToken(ctx, token2, userID, time.Hour)

	err = repo.RevokeAllUserTokens(ctx, userID)
	assert.NoError(t, err)

	_, err1 := repo.ValidateRefreshToken(ctx, token1)
	_, err2 := repo.ValidateRefreshToken(ctx, token2)
	assert.Error(t, err1)
	assert.Error(t, err2)

	// 6. TTL Expiration via MiniRedis Fast-Forward
	ttlToken := "token-ttl-test"
	_ = repo.StoreRefreshToken(ctx, ttlToken, userID, 5*time.Second)
	mr.FastForward(6 * time.Second)

	_, err = repo.ValidateRefreshToken(ctx, ttlToken)
	assert.Error(t, err)
}

func TestRedisTokenRepository_NilClient(t *testing.T) {
	repo := persistence.NewRedisTokenRepository(nil)
	ctx := context.Background()

	assert.Error(t, repo.StoreRefreshToken(ctx, "t", "u", time.Hour))
	_, err := repo.ValidateRefreshToken(ctx, "t")
	assert.Error(t, err)
	assert.Error(t, repo.RevokeRefreshToken(ctx, "t"))
	assert.Error(t, repo.RevokeAllUserTokens(ctx, "u"))
}
