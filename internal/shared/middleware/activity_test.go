package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockUserActivityRepo struct {
	updateCount int64
	lastUpdated time.Time
	lastUserID  string
}

func (m *mockUserActivityRepo) Create(ctx context.Context, user *entity.User) error { return nil }
func (m *mockUserActivityRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserActivityRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserActivityRepo) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *mockUserActivityRepo) UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error {
	return nil
}
func (m *mockUserActivityRepo) UpdateLastActivityAt(ctx context.Context, userID string, at time.Time) error {
	atomic.AddInt64(&m.updateCount, 1)
	m.lastUpdated = at
	m.lastUserID = userID
	return nil
}

func setupActivityRouter(repo *mockUserActivityRepo, rdb *redis.Client, interval time.Duration, enabled bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/public", func(c *gin.Context) {
		c.String(http.StatusOK, "public")
	})

	protected := r.Group("/api")
	// Simulate auth setting user_id
	protected.Use(func(c *gin.Context) {
		uid := c.GetHeader("X-Test-User-ID")
		if uid != "" {
			c.Set(middleware.ContextUserIDKey, uid)
		}
		c.Next()
	})
	protected.Use(middleware.UserActivityTracker(repo, rdb, interval, enabled))
	protected.GET("/resource", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	return r
}

func TestUserActivityTracker_Disabled(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	repo := &mockUserActivityRepo{}
	r := setupActivityRouter(repo, rdb, 5*time.Minute, false)

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.Header.Set("X-Test-User-ID", "user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(0), atomic.LoadInt64(&repo.updateCount))
}

func TestUserActivityTracker_Unauthenticated(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	repo := &mockUserActivityRepo{}
	r := setupActivityRouter(repo, rdb, 5*time.Minute, true)

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	// No X-Test-User-ID header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(0), atomic.LoadInt64(&repo.updateCount))
}

func TestUserActivityTracker_Throttling_WithRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	repo := &mockUserActivityRepo{}
	interval := 5 * time.Minute
	r := setupActivityRouter(repo, rdb, interval, true)

	// 1. First Request -> Triggers DB write and sets Redis key
	req1 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req1.Header.Set("X-Test-User-ID", "user-100")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, int64(1), atomic.LoadInt64(&repo.updateCount))
	assert.Equal(t, "user-100", repo.lastUserID)

	// Check Redis key exists
	exists := mr.Exists("user:activity:user-100")
	assert.True(t, exists)

	// 2. Second Request within interval -> Throttled, DB not updated
	req2 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req2.Header.Set("X-Test-User-ID", "user-100")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, int64(1), atomic.LoadInt64(&repo.updateCount)) // Still 1

	// 3. Fast-forward past interval -> Next request triggers DB update
	mr.FastForward(6 * time.Minute)

	req3 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req3.Header.Set("X-Test-User-ID", "user-100")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Equal(t, int64(2), atomic.LoadInt64(&repo.updateCount)) // Incremented to 2
}

func TestUserActivityTracker_FallbackWithoutRedis(t *testing.T) {
	repo := &mockUserActivityRepo{}
	interval := 100 * time.Millisecond
	r := setupActivityRouter(repo, nil, interval, true)

	// 1. First Request -> Updates DB via in-memory fallback
	req1 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req1.Header.Set("X-Test-User-ID", "fallback-user")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, int64(1), atomic.LoadInt64(&repo.updateCount))

	// 2. Immediate request -> Throttled
	req2 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req2.Header.Set("X-Test-User-ID", "fallback-user")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, int64(1), atomic.LoadInt64(&repo.updateCount))

	// 3. Wait for interval -> Next request updates
	time.Sleep(150 * time.Millisecond)

	req3 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req3.Header.Set("X-Test-User-ID", "fallback-user")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Equal(t, int64(2), atomic.LoadInt64(&repo.updateCount))
}
