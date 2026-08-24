package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS_Headers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.CORS([]string{"https://example.com"}))
	r.OPTIONS("/api/v1/articles", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.GET("/api/v1/articles", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	t.Run("Preflight OPTIONS request from allowed origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/articles", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization, Idempotency-Key")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "idempotency-key")
	})

	t.Run("Exposed response headers include RateLimit and IdempotentReplay", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles", nil)
		req.Header.Set("Origin", "https://example.com")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		exposed := strings.ToLower(w.Header().Get("Access-Control-Expose-Headers"))
		assert.Contains(t, exposed, "x-idempotent-replay")
		assert.Contains(t, exposed, "x-ratelimit-limit")
	})
}
