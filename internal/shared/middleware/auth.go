package middleware

import (
	"context"
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "
)

// AuthMiddleware creates a Gin middleware that validates JWT Bearer tokens.
func AuthMiddleware(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(HeaderAuthorization)
		if authHeader == "" {
			response.Error(c, appErrors.NewUnauthorizedError("authorization header is required"))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Error(c, appErrors.NewUnauthorizedError("invalid authorization header format. Expected 'Bearer <token>'"))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := tokenService.ValidateAccessToken(tokenStr)
		if err != nil {
			response.Error(c, appErrors.NewUnauthorizedError("invalid or expired access token", err))
			c.Abort()
			return
		}

		// Inject into Gin context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmailKey, claims.Email)

		// Inject into request context for structured logging
		ctx := context.WithValue(c.Request.Context(), logger.UserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetAuthenticatedUserID retrieves the user ID from Gin context.
func GetAuthenticatedUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}
