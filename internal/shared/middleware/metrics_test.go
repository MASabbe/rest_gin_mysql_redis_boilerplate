package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.Metrics())
	r.GET("/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "item")
	})
	r.GET("/error", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	t.Run("Matched route records template path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/items/123-abc-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unmatched route records /unmatched", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/non-existent-endpoint-404", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Server error records error metric", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/error", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
