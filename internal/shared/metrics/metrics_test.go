package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Recording_And_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Record sample metrics
	metrics.RecordHTTPRequest("GET", "/api/v1/articles", "200", 0.042)
	metrics.RecordHTTPError("POST", "/api/v1/auth/login", "UNAUTHORIZED")
	metrics.RecordAuthFailure("invalid_credentials")
	metrics.RecordRBACDenial("article:delete")
	metrics.RecordRateLimitRejection("auth")
	metrics.RecordCacheHit("rbac")
	metrics.RecordCacheMiss("rbac")

	r := gin.New()
	r.GET("/metrics", metrics.Handler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "app_http_requests_total")
	assert.Contains(t, body, "app_http_request_duration_seconds")
	assert.Contains(t, body, "app_http_errors_total")
	assert.Contains(t, body, "app_auth_failures_total")
	assert.Contains(t, body, "app_rbac_denials_total")
	assert.Contains(t, body, "app_ratelimit_rejections_total")
	assert.Contains(t, body, "app_cache_hits_total")
	assert.Contains(t, body, "app_cache_misses_total")
}

func TestMetrics_RegisterDBStats_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() {
		metrics.RegisterDBStats(nil)
	})
}
