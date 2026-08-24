package middleware

import (
	"strconv"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
	"github.com/gin-gonic/gin"
)

// Metrics records Prometheus HTTP metrics using route templates to prevent unbounded cardinality.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// Use normalized route template (e.g. "/api/v1/articles/:id") instead of raw path with IDs
		path := c.FullPath()
		if path == "" {
			path = "/unmatched"
		}

		metrics.RecordHTTPRequest(method, path, status, duration)

		if c.Writer.Status() >= 400 {
			errorType := "CLIENT_ERROR"
			if c.Writer.Status() >= 500 {
				errorType = "SERVER_ERROR"
			}
			metrics.RecordHTTPError(method, path, errorType)
		}
	}
}
