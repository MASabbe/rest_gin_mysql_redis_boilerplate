package middleware

import (
	"fmt"
	"runtime/debug"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Recovery recovers from panics and returns a clean 500 JSON response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				err := fmt.Errorf("panic recovered: %v", r)

				logger.WithContext(c.Request.Context()).Error("HTTP Request Panic",
					"error", err,
					"stack", stack,
				)

				response.Error(c, appErrors.NewInternalError("An unexpected server error occurred", err))
				c.Abort()
			}
		}()
		c.Next()
	}
}
