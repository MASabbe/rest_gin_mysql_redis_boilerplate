package handler

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http/request"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http/response"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type UserRoleHandler struct {
	rbacService application.RBACService
}

// NewUserRoleHandler creates a new UserRoleHandler instance.
func NewUserRoleHandler(rbacService application.RBACService) *UserRoleHandler {
	return &UserRoleHandler{rbacService: rbacService}
}

// AssignUserRoles handles PUT /api/v1/users/:id/roles
func (h *UserRoleHandler) AssignUserRoles(c *gin.Context) {
	targetUserID := c.Param("id")
	var req request.AssignUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "role_ids",
			Message: err.Error(),
		}))
		return
	}

	err := h.rbacService.AssignRolesToUser(c.Request.Context(), command.AssignUserRolesCommand{
		UserID:  targetUserID,
		RoleIDs: req.RoleIDs,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "User roles updated successfully", nil)
}

// GetUserRoles handles GET /api/v1/users/:id/roles
func (h *UserRoleHandler) GetUserRoles(c *gin.Context) {
	targetUserID := c.Param("id")
	result, err := h.rbacService.GetUserRoles(c.Request.Context(), query.GetUserRolesQuery{UserID: targetUserID})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "User roles retrieved successfully", response.FromUserRolesDTO(result))
}

// GetUserPermissions handles GET /api/v1/users/:id/permissions
func (h *UserRoleHandler) GetUserPermissions(c *gin.Context) {
	targetUserID := c.Param("id")
	result, err := h.rbacService.GetUserPermissionsDTO(c.Request.Context(), query.GetUserPermissionsQuery{UserID: targetUserID})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "User permissions retrieved successfully", response.FromUserPermissionsDTO(result))
}

// GetMyPermissions handles GET /api/v1/auth/me/permissions (for current authenticated user)
func (h *UserRoleHandler) GetMyPermissions(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		sharedResponse.Error(c, appErrors.NewUnauthorizedError("unauthorized"))
		return
	}

	result, err := h.rbacService.GetUserPermissionsDTO(c.Request.Context(), query.GetUserPermissionsQuery{UserID: userID})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "User permissions retrieved successfully", response.FromUserPermissionsDTO(result))
}
