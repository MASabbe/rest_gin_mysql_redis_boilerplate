package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	RedisUserActivityPrefix = "user:activity:"
)

// inMemoryThrottle provides a bounded local fallback when Redis is unavailable.
type inMemoryThrottle struct {
	mu      sync.RWMutex
	records map[string]time.Time
}

var localThrottle = &inMemoryThrottle{
	records: make(map[string]time.Time),
}

func (m *inMemoryThrottle) shouldUpdate(userID string, interval time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	lastUpdate, exists := m.records[userID]
	if exists && now.Sub(lastUpdate) < interval {
		return false
	}

	// Clean up stale entries if cache grows beyond threshold (prevent unbounded memory)
	if len(m.records) > 10000 {
		for id, t := range m.records {
			if now.Sub(t) > interval*2 {
				delete(m.records, id)
			}
		}
	}

	m.records[userID] = now
	return true
}

// UserActivityTracker creates a Gin middleware that records the user's last activity timestamp.
// It uses Redis to throttle MySQL database writes based on the configured update interval.
func UserActivityTracker(
	userRepo repository.UserRepository,
	redisClient *redis.Client,
	interval time.Duration,
	enabled bool,
) gin.HandlerFunc {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	return func(c *gin.Context) {
		// 1. Check if activity tracking is enabled
		if !enabled {
			c.Next()
			return
		}

		// 2. Extract authenticated user ID from context
		userID, exists := GetAuthenticatedUserID(c)
		if !exists || userID == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		shouldUpdateDB := false

		// 3. Throttle check using Redis SetNX
		if redisClient != nil {
			redisKey := RedisUserActivityPrefix + userID
			// SetNX sets the key only if it does not already exist
			set, err := redisClient.SetNX(ctx, redisKey, "1", interval).Result()
			if err != nil {
				// Redis failure -> Log warning and use in-process bounded fallback throttle
				logger.WithContext(ctx).Warn("Redis activity throttle unavailable, using in-memory fallback",
					slog.String("user_id", userID),
					slog.String("error", err.Error()),
				)
				shouldUpdateDB = localThrottle.shouldUpdate(userID, interval)
			} else {
				shouldUpdateDB = set
			}
		} else {
			// No Redis configured -> Use in-memory fallback throttle
			shouldUpdateDB = localThrottle.shouldUpdate(userID, interval)
		}

		// 4. Update database timestamp if eligible
		if shouldUpdateDB && userRepo != nil {
			now := time.Now().UTC()
			if err := userRepo.UpdateLastActivityAt(ctx, userID, now); err != nil {
				logger.WithContext(ctx).Warn("Failed to update last_activity_at for user",
					slog.String("user_id", userID),
					slog.String("error", err.Error()),
				)
			}
		}

		c.Next()
	}
}
