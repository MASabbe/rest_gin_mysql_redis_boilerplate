package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository creates a new RedisTokenRepository instance.
func NewRedisTokenRepository(client *redis.Client) repository.TokenRepository {
	return &RedisTokenRepository{client: client}
}

func (r *RedisTokenRepository) formatTokenKey(tokenID string) string {
	return fmt.Sprintf("auth:refresh:%s", tokenID)
}

func (r *RedisTokenRepository) formatUserTokensKey(userID string) string {
	return fmt.Sprintf("auth:user_tokens:%s", userID)
}

// StoreRefreshToken saves the refresh token ID with associated user ID and TTL.
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, tokenID string, userID string, ttl time.Duration) error {
	if r.client == nil {
		return appErrors.NewInfrastructureError("redis client is uninitialized")
	}

	tokenKey := r.formatTokenKey(tokenID)
	userTokensKey := r.formatUserTokensKey(userID)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, tokenKey, userID, ttl)
	pipe.SAdd(ctx, userTokensKey, tokenID)
	pipe.Expire(ctx, userTokensKey, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return appErrors.NewInfrastructureError("failed to store refresh token in redis", err)
	}

	return nil
}

// ValidateRefreshToken checks if the refresh token exists and returns the associated userID.
func (r *RedisTokenRepository) ValidateRefreshToken(ctx context.Context, tokenID string) (string, error) {
	if r.client == nil {
		return "", appErrors.NewInfrastructureError("redis client is uninitialized")
	}

	tokenKey := r.formatTokenKey(tokenID)
	userID, err := r.client.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", appErrors.NewUnauthorizedError("refresh token has been revoked or expired")
		}
		return "", appErrors.NewInfrastructureError("failed to validate refresh token from redis", err)
	}

	return userID, nil
}

// RevokeRefreshToken removes the refresh token key from Redis.
func (r *RedisTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	if r.client == nil {
		return appErrors.NewInfrastructureError("redis client is uninitialized")
	}

	tokenKey := r.formatTokenKey(tokenID)
	err := r.client.Del(ctx, tokenKey).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return appErrors.NewInfrastructureError("failed to revoke refresh token", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all active refresh tokens for the given user ID.
func (r *RedisTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if r.client == nil {
		return appErrors.NewInfrastructureError("redis client is uninitialized")
	}

	userTokensKey := r.formatUserTokensKey(userID)
	tokenIDs, err := r.client.SMembers(ctx, userTokensKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return appErrors.NewInfrastructureError("failed to retrieve user tokens", err)
	}

	if len(tokenIDs) > 0 {
		pipe := r.client.Pipeline()
		for _, tid := range tokenIDs {
			pipe.Del(ctx, r.formatTokenKey(tid))
		}
		pipe.Del(ctx, userTokensKey)
		_, _ = pipe.Exec(ctx)
	}

	return nil
}
