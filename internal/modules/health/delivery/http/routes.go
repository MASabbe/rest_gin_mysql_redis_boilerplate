package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the health check routes to the Gin engine or router group.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	health := router.Group("/health")
	{
		health.GET("", h.Health)
		health.GET("/live", h.Live)
		health.GET("/ready", h.Ready)
	}
}
