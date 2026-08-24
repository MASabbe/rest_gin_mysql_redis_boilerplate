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

type PermissionHandler struct {
	rbacService application.RBACService
}

// NewPermissionHandler creates a new PermissionHandler instance.
func NewPermissionHandler(rbacService application.RBACService) *PermissionHandler {
	return &PermissionHandler{rbacService: rbacService}
}

// Create handles POST /api/v1/permissions
func (h *PermissionHandler) Create(c *gin.Context) {
	var req request.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.rbacService.CreatePermission(c.Request.Context(), command.CreatePermissionCommand{
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Created(c, "Permission created successfully", response.FromPermissionDTO(result))
}

// GetByID handles GET /api/v1/permissions/:id
func (h *PermissionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.rbacService.GetPermission(c.Request.Context(), query.GetPermissionByIDQuery{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Permission retrieved successfully", response.FromPermissionDTO(result))
}

// List handles GET /api/v1/permissions
func (h *PermissionHandler) List(c *gin.Context) {
	p := pagination.Extract(c)
	list, meta, err := h.rbacService.ListPermissions(c.Request.Context(), query.ListPermissionsQuery{Pagination: p})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Success(c, 200, "Permissions retrieved successfully", response.FromPermissionDTOList(list), meta)
}

// Update handles PUT /api/v1/permissions/:id
func (h *PermissionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req request.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.rbacService.UpdatePermission(c.Request.Context(), command.UpdatePermissionCommand{
		ID:          id,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Permission updated successfully", response.FromPermissionDTO(result))
}

// Delete handles DELETE /api/v1/permissions/:id
func (h *PermissionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.rbacService.DeletePermission(c.Request.Context(), command.DeletePermissionCommand{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Permission deleted successfully", nil)
}
