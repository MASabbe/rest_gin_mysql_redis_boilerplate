package http

import (
	"net/http"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service application.HealthService
}

func NewHandler(service application.HealthService) *Handler {
	return &Handler{service: service}
}

// Live handles GET /health/live (Liveness probe)
func (h *Handler) Live(c *gin.Context) {
	report := h.service.CheckLiveness(c.Request.Context())
	response.OK(c, "Service is alive", report)
}

// Ready handles GET /health/ready (Readiness probe)
func (h *Handler) Ready(c *gin.Context) {
	report := h.service.CheckReadiness(c.Request.Context())
	if report.Status == domain.StatusDOWN {
		response.Success(c, http.StatusServiceUnavailable, "Service is not ready", report)
		return
	}
	response.OK(c, "Service is ready", report)
}

// Health handles GET /health (Overall health)
func (h *Handler) Health(c *gin.Context) {
	report := h.service.CheckOverall(c.Request.Context())
	if report.Status == domain.StatusDOWN {
		response.Success(c, http.StatusServiceUnavailable, "Service health check failed", report)
		return
	}
	response.OK(c, "Service health status", report)
}
