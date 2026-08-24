package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS initializes and returns CORS middleware with secure defaults.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-Trace-ID", "traceparent", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-Trace-ID", "traceparent", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After", "X-Idempotent-Replay"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
	}

	return cors.New(corsConfig)
}
