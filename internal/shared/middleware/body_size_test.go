package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMaxBodySize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.MaxBodySize(100)) // Max 100 bytes
	r.POST("/upload", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	t.Run("Payload under limit passes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("short payload"))
		req.Header.Set("Content-Type", "text/plain")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Payload exceeding limit returns 413", func(t *testing.T) {
		w := httptest.NewRecorder()
		hugePayload := strings.Repeat("A", 200)
		req, _ := http.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString(hugePayload))
		req.Header.Set("Content-Type", "text/plain")
		r.ServeHTTP(w, req)

		assert.Equal(t, 413, w.Code)
	})
}
