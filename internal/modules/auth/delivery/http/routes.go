package http

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/handler"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers auth endpoints to the given Gin router group.
func RegisterRoutes(
	apiGroup *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	tokenService service.TokenService,
	activityMiddleware ...gin.HandlerFunc,
) {
	auth := apiGroup.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)

		// Protected auth routes
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(tokenService))
		for _, m := range activityMiddleware {
			if m != nil {
				protected.Use(m)
			}
		}
		{
			protected.GET("/me", authHandler.Me)
		}
	}
}
