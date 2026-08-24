package http

import (
	authService "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http/handler"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers RBAC endpoints to the provided Gin router group with Authentication, Activity Tracking, and Permission guards.
func RegisterRoutes(
	apiGroup *gin.RouterGroup,
	tokenService authService.TokenService,
	rbacService application.RBACService,
	activityMiddleware ...gin.HandlerFunc,
) {
	roleHdlr := handler.NewRoleHandler(rbacService)
	permHdlr := handler.NewPermissionHandler(rbacService)
	userRoleHdlr := handler.NewUserRoleHandler(rbacService)

	// Protected RBAC Root Group (requires valid Bearer token)
	rbac := apiGroup.Group("")
	rbac.Use(middleware.AuthMiddleware(tokenService))
	for _, m := range activityMiddleware {
		if m != nil {
			rbac.Use(m)
		}
	}
	{
		// 1. Roles Management Endpoints
		roles := rbac.Group("/roles")
		{
			roles.POST("", middleware.RequirePermission(rbacService, "role:create"), roleHdlr.Create)
			roles.GET("", middleware.RequirePermission(rbacService, "role:read"), roleHdlr.List)
			roles.GET("/:id", middleware.RequirePermission(rbacService, "role:read"), roleHdlr.GetByID)
			roles.PUT("/:id", middleware.RequirePermission(rbacService, "role:update"), roleHdlr.Update)
			roles.DELETE("/:id", middleware.RequirePermission(rbacService, "role:delete"), roleHdlr.Delete)
			roles.PUT("/:id/permissions", middleware.RequirePermission(rbacService, "role:update"), roleHdlr.AssignPermissions)
			roles.GET("/:id/permissions", middleware.RequirePermission(rbacService, "role:read"), roleHdlr.GetRolePermissions)
		}

		// 2. Permissions Management Endpoints
		permissions := rbac.Group("/permissions")
		{
			permissions.POST("", middleware.RequirePermission(rbacService, "permission:create"), permHdlr.Create)
			permissions.GET("", middleware.RequirePermission(rbacService, "permission:read"), permHdlr.List)
			permissions.GET("/:id", middleware.RequirePermission(rbacService, "permission:read"), permHdlr.GetByID)
			permissions.PUT("/:id", middleware.RequirePermission(rbacService, "permission:update"), permHdlr.Update)
			permissions.DELETE("/:id", middleware.RequirePermission(rbacService, "permission:delete"), permHdlr.Delete)
		}

		// 3. User Role & Permission Endpoints
		users := rbac.Group("/users")
		{
			users.PUT("/:id/roles", middleware.RequirePermission(rbacService, "user:assign-role"), userRoleHdlr.AssignUserRoles)
			users.GET("/:id/roles", middleware.RequirePermission(rbacService, "role:read"), userRoleHdlr.GetUserRoles)
			users.GET("/:id/permissions", middleware.RequirePermission(rbacService, "permission:read"), userRoleHdlr.GetUserPermissions)
		}

		// 4. Current Authenticated User Permissions
		rbac.GET("/auth/me/permissions", userRoleHdlr.GetMyPermissions)
	}
}
