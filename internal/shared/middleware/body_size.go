package middleware

import (
	"net/http"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// MaxBodySize limits the maximum allowed request body size in bytes.
// If the payload exceeds maxBytes, http.MaxBytesReader causes a read error and responds with 413 Payload Too Large.
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024 // Default 2MB
	}

	return func(c *gin.Context) {
		if c.Request.Body != nil && c.Request.ContentLength > maxBytes {
			sharedResponse.Error(c, appErrors.NewPayloadTooLargeError("payload exceeds maximum allowed limit"))
			c.Abort()
			return
		}

		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}

		c.Next()
	}
}
