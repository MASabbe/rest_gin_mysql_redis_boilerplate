package middleware

import (
	"log/slog"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/gin-gonic/gin"
)

// Logger logs HTTP request details using structured slog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		attrs := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if errorMessage != "" {
			attrs = append(attrs, slog.String("error", errorMessage))
		}

		log := logger.WithContext(c.Request.Context())
		if status >= 500 {
			log.Error("HTTP Request Error", attrs...)
		} else if status >= 400 {
			log.Warn("HTTP Request Client Warning", attrs...)
		} else {
			log.Info("HTTP Request", attrs...)
		}
	}
}
