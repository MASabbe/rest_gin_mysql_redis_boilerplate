package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout middleware attaches a context timeout to the request context.
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if duration <= 0 {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
