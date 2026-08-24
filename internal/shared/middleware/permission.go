package middleware

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/service"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// RequirePermission enforces that the authenticated user possesses the specified permission.
func RequirePermission(authzService service.AuthorizationService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetAuthenticatedUserID(c)
		if !ok || userID == "" {
			response.Error(c, appErrors.NewUnauthorizedError("authentication required"))
			c.Abort()
			return
		}

		hasPerm, err := authzService.HasPermission(c.Request.Context(), userID, permission)
		if err != nil {
			response.Error(c, appErrors.NewInternalError("failed to authorize request", err))
			c.Abort()
			return
		}

		if !hasPerm {
			metrics.RecordRBACDenial(permission)
			response.Error(c, appErrors.NewForbiddenError("Forbidden"))
			c.Abort()
			return
		}

		c.Next()
	}
}
