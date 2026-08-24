package http

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/delivery/http/handler"
	authService "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	rbacService "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers article endpoints with authentication and permission protection.
func RegisterRoutes(
	apiGroup *gin.RouterGroup,
	tokenService authService.TokenService,
	authzService rbacService.AuthorizationService,
	articleService application.ArticleService,
) {
	hdlr := handler.NewArticleHandler(articleService)

	articles := apiGroup.Group("/articles")
	{
		// Public or authenticated list & get
		articles.GET("", hdlr.List)
		articles.GET("/:id", hdlr.GetByID)
		articles.GET("/slug/:slug", hdlr.GetBySlug)

		// Protected endpoints
		protected := articles.Group("")
		protected.Use(middleware.AuthMiddleware(tokenService))
		{
			protected.POST("", middleware.RequirePermission(authzService, "article:create"), hdlr.Create)
			protected.PUT("/:id", middleware.RequirePermission(authzService, "article:update"), hdlr.Update)
			protected.DELETE("/:id", middleware.RequirePermission(authzService, "article:delete"), hdlr.Delete)
		}
	}
}
