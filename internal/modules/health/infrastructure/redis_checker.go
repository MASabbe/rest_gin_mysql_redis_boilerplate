package infrastructure

import (
	"context"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/redis"
)

type RedisChecker struct {
	redis *redis.Redis
}

// NewRedisChecker creates a Checker for Redis.
func NewRedisChecker(redis *redis.Redis) domain.Checker {
	return &RedisChecker{redis: redis}
}

func (c *RedisChecker) Name() string {
	return "redis"
}

func (c *RedisChecker) Check(ctx context.Context) domain.ComponentCheck {
	start := time.Now()
	if c.redis == nil {
		return domain.ComponentCheck{
			Name:      c.Name(),
			Status:    domain.StatusDOWN,
			Error:     "redis client is not configured",
			Latency:   time.Since(start),
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}
	}

	err := c.redis.Ping(ctx)
	latency := time.Since(start)
	latencyMs := float64(latency.Microseconds()) / 1000.0

	if err != nil {
		return domain.ComponentCheck{
			Name:      c.Name(),
			Status:    domain.StatusDOWN,
			Error:     err.Error(),
			Latency:   latency,
			LatencyMs: latencyMs,
		}
	}

	return domain.ComponentCheck{
		Name:      c.Name(),
		Status:    domain.StatusUP,
		Details:   "cache server responsive",
		Latency:   latency,
		LatencyMs: latencyMs,
	}
}
