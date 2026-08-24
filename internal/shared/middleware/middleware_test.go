package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_RequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/test-req-id", func(c *gin.Context) {
		reqID := c.GetString("request_id")
		assert.NotEmpty(t, reqID)
		c.Status(http.StatusOK)
	})

	// Case 1: Auto-generated
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test-req-id", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(middleware.HeaderXRequestID))
	assert.NotEmpty(t, w.Header().Get(middleware.HeaderXTraceID))
	assert.NotEmpty(t, w.Header().Get(middleware.HeaderTraceParent))

	// Case 2: Custom Header preserved
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test-req-id", nil)
	req2.Header.Set(middleware.HeaderXRequestID, "custom-id-999")
	req2.Header.Set(middleware.HeaderXTraceID, "4bf92f3577b34da6a3ce929d0e0e4736")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, "custom-id-999", w2.Header().Get(middleware.HeaderXRequestID))
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", w2.Header().Get(middleware.HeaderXTraceID))
}

func TestMiddleware_Recovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("something went terribly wrong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMiddleware_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS([]string{"*"}))
	r.OPTIONS("/test-cors", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test-cors", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestMiddleware_Timeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Timeout(50 * time.Millisecond))
	r.GET("/fast", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/fast", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_Logger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Logger())
	r.GET("/log-test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/log-warn", func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})
	r.GET("/log-err", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	for _, path := range []string{"/log-test?query=val", "/log-warn", "/log-err"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
	}
}
