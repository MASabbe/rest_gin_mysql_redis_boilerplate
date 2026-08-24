package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter returns a distributed rate limiting middleware backed by Redis.
// limit: maximum requests allowed within window (1 minute).
// tier: descriptive label for the scope (e.g. "general", "auth").
func RateLimiter(redisClient *redis.Client, limit int, tier string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit <= 0 {
			c.Next()
			return
		}

		if redisClient == nil {
			// Graceful fallback if Redis is not configured
			c.Next()
			return
		}

		// Identify client by authenticated user ID if present, otherwise by IP
		identifier := c.ClientIP()
		if userID, ok := GetAuthenticatedUserID(c); ok && userID != "" {
			identifier = "user:" + userID
		}

		now := time.Now().UTC()
		window := now.Unix() / 60
		windowReset := (window + 1) * 60
		secondsRemaining := windowReset - now.Unix()
		if secondsRemaining < 1 {
			secondsRemaining = 1
		}

		cacheKey := fmt.Sprintf("ratelimit:%s:%s:%d", tier, identifier, window)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		pipe := redisClient.Pipeline()
		incr := pipe.Incr(ctx, cacheKey)
		pipe.Expire(ctx, cacheKey, 65*time.Second)
		_, err := pipe.Exec(ctx)

		if err != nil {
			// Fail-open with degraded log warning
			logger.Get().Warn("Rate limiter Redis degraded, failing open",
				slog.String("tier", tier),
				slog.String("error", err.Error()),
			)
			c.Next()
			return
		}

		currentCount := incr.Val()
		remaining := int64(limit) - currentCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(windowReset, 10))

		if currentCount > int64(limit) {
			metrics.RecordRateLimitRejection(tier)
			c.Header("Retry-After", strconv.FormatInt(secondsRemaining, 10))
			sharedResponse.Error(c, appErrors.NewRateLimitExceededError(
				fmt.Sprintf("rate limit exceeded, retry after %d seconds", secondsRemaining),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
