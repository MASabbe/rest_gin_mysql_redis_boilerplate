package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_FallbackAndPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Nil client passes through safely", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.RateLimiter(nil, 10, "general"))
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Degraded redis client fails open without crashing", func(t *testing.T) {
		// Pointing to a closed/offline port
		badClient := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:1",
		})
		defer badClient.Close()

		r := gin.New()
		r.Use(middleware.RateLimiter(badClient, 10, "auth"))
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
