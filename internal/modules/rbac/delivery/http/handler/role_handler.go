package handler

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http/request"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http/response"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	rbacService application.RBACService
}

// NewRoleHandler creates a new RoleHandler instance.
func NewRoleHandler(rbacService application.RBACService) *RoleHandler {
	return &RoleHandler{rbacService: rbacService}
}

// Create handles POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req request.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.rbacService.CreateRole(c.Request.Context(), command.CreateRoleCommand{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Created(c, "Role created successfully", response.FromRoleDTO(result))
}

// GetByID handles GET /api/v1/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.rbacService.GetRole(c.Request.Context(), query.GetRoleByIDQuery{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Role retrieved successfully", response.FromRoleDTO(result))
}

// List handles GET /api/v1/roles
func (h *RoleHandler) List(c *gin.Context) {
	p := pagination.Extract(c)
	list, meta, err := h.rbacService.ListRoles(c.Request.Context(), query.ListRolesQuery{Pagination: p})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Success(c, 200, "Roles retrieved successfully", response.FromRoleDTOList(list), meta)
}

// Update handles PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.rbacService.UpdateRole(c.Request.Context(), command.UpdateRoleCommand{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Role updated successfully", response.FromRoleDTO(result))
}

// Delete handles DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.rbacService.DeleteRole(c.Request.Context(), command.DeleteRoleCommand{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Role deleted successfully", nil)
}

// AssignPermissions handles PUT /api/v1/roles/:id/permissions
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	roleID := c.Param("id")
	var req request.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "permission_ids",
			Message: err.Error(),
		}))
		return
	}

	err := h.rbacService.AssignPermissionsToRole(c.Request.Context(), command.AssignRolePermissionsCommand{
		RoleID:        roleID,
		PermissionIDs: req.PermissionIDs,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Role permissions updated successfully", nil)
}

// GetRolePermissions handles GET /api/v1/roles/:id/permissions
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	result, err := h.rbacService.GetRolePermissions(c.Request.Context(), query.GetRolePermissionsQuery{RoleID: roleID})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Role permissions retrieved successfully", response.FromRoleWithPermissionsDTO(result))
}
