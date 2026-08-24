package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIdempotency_Passthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GET request bypasses idempotency", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.Idempotency(nil, 24*time.Hour))
		r.GET("/items", func(c *gin.Context) {
			c.String(http.StatusOK, "list items")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/items", nil)
		req.Header.Set("Idempotency-Key", "test-key-1")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "list items", w.Body.String())
	})

	t.Run("POST request without key passes through", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.Idempotency(nil, 24*time.Hour))
		r.POST("/items", func(c *gin.Context) {
			c.String(http.StatusCreated, "created item")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/items", bytes.NewBufferString(`{"name":"foo"}`))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "created item", w.Body.String())
	})
}
