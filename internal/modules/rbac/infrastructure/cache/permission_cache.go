package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PermissionCache defines caching contract for effective user permissions.
type PermissionCache interface {
	Get(ctx context.Context, userID string) ([]string, bool, error)
	Set(ctx context.Context, userID string, permissions []string, ttl time.Duration) error
	InvalidateUser(ctx context.Context, userID string) error
	InvalidateUsers(ctx context.Context, userIDs []string) error
}

type RedisPermissionCache struct {
	client *redis.Client
}

// NewRedisPermissionCache creates a new RedisPermissionCache.
func NewRedisPermissionCache(client *redis.Client) PermissionCache {
	return &RedisPermissionCache{client: client}
}

func (c *RedisPermissionCache) formatKey(userID string) string {
	return fmt.Sprintf("rbac:user:%s:permissions", userID)
}

// Get retrieves cached permissions for a user. Returns (perms, hit, error).
func (c *RedisPermissionCache) Get(ctx context.Context, userID string) ([]string, bool, error) {
	if c.client == nil {
		return nil, false, nil // Graceful fallback
	}

	key := c.formatKey(userID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil // Cache miss
		}
		return nil, false, err // Redis error, will fallback to DB
	}

	var permissions []string
	if err := json.Unmarshal([]byte(val), &permissions); err != nil {
		return nil, false, err
	}

	return permissions, true, nil
}

// Set stores permissions in Redis with the given TTL.
func (c *RedisPermissionCache) Set(ctx context.Context, userID string, permissions []string, ttl time.Duration) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions for cache: %w", err)
	}

	key := c.formatKey(userID)
	return c.client.Set(ctx, key, data, ttl).Err()
}

// InvalidateUser evicts a specific user's cached permissions.
func (c *RedisPermissionCache) InvalidateUser(ctx context.Context, userID string) error {
	if c.client == nil {
		return nil
	}

	key := c.formatKey(userID)
	return c.client.Del(ctx, key).Err()
}

// InvalidateUsers evicts permissions for multiple users in a single pipeline.
func (c *RedisPermissionCache) InvalidateUsers(ctx context.Context, userIDs []string) error {
	if c.client == nil || len(userIDs) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	for _, uid := range userIDs {
		pipe.Del(ctx, c.formatKey(uid))
	}
	_, err := pipe.Exec(ctx)
	return err
}
