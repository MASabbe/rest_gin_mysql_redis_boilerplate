package application

import (
	"context"
	"fmt"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/dto"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/cache"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

const (
	DefaultPermissionCacheTTL = 1 * time.Hour
)

// RBACService combines RBAC administrative capabilities and Authorization evaluation.
type RBACService interface {
	service.AuthorizationService

	// Role Management
	CreateRole(ctx context.Context, cmd command.CreateRoleCommand) (*dto.RoleDTO, error)
	GetRole(ctx context.Context, q query.GetRoleByIDQuery) (*dto.RoleDTO, error)
	ListRoles(ctx context.Context, q query.ListRolesQuery) ([]dto.RoleDTO, pagination.Meta, error)
	UpdateRole(ctx context.Context, cmd command.UpdateRoleCommand) (*dto.RoleDTO, error)
	DeleteRole(ctx context.Context, cmd command.DeleteRoleCommand) error
	AssignPermissionsToRole(ctx context.Context, cmd command.AssignRolePermissionsCommand) error
	GetRolePermissions(ctx context.Context, q query.GetRolePermissionsQuery) (*dto.RoleWithPermissionsDTO, error)

	// Permission Management
	CreatePermission(ctx context.Context, cmd command.CreatePermissionCommand) (*dto.PermissionDTO, error)
	GetPermission(ctx context.Context, q query.GetPermissionByIDQuery) (*dto.PermissionDTO, error)
	ListPermissions(ctx context.Context, q query.ListPermissionsQuery) ([]dto.PermissionDTO, pagination.Meta, error)
	UpdatePermission(ctx context.Context, cmd command.UpdatePermissionCommand) (*dto.PermissionDTO, error)
	DeletePermission(ctx context.Context, cmd command.DeletePermissionCommand) error

	// User RBAC Management
	AssignRolesToUser(ctx context.Context, cmd command.AssignUserRolesCommand) error
	GetUserRoles(ctx context.Context, q query.GetUserRolesQuery) (*dto.UserRolesDTO, error)
	GetUserPermissionsDTO(ctx context.Context, q query.GetUserPermissionsQuery) (*dto.UserPermissionsDTO, error)
}

type rbacService struct {
	roleRepo        repository.RoleRepository
	permissionRepo  repository.PermissionRepository
	permissionCache cache.PermissionCache
	cacheTTL        time.Duration
}

// NewRBACService creates a new RBACService instance.
func NewRBACService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	permissionCache cache.PermissionCache,
	cacheTTL ...time.Duration,
) RBACService {
	ttl := DefaultPermissionCacheTTL
	if len(cacheTTL) > 0 && cacheTTL[0] > 0 {
		ttl = cacheTTL[0]
	}

	return &rbacService{
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		permissionCache: permissionCache,
		cacheTTL:        ttl,
	}
}

// --- Authorization Service Implementation ---

func (s *rbacService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	if userID == "" || permission == "" {
		return false, nil
	}

	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

func (s *rbacService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return []string{}, nil
	}

	// 1. Try Cache
	if s.permissionCache != nil {
		if cached, hit, _ := s.permissionCache.Get(ctx, userID); hit {
			return cached, nil
		}
	}

	// 2. Query MySQL / DB
	perms, err := s.permissionRepo.FindEffectivePermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. Populate Cache
	if s.permissionCache != nil {
		_ = s.permissionCache.Set(ctx, userID, perms, s.cacheTTL)
	}

	return perms, nil
}

// --- Role Management ---

func (s *rbacService) CreateRole(ctx context.Context, cmd command.CreateRoleCommand) (*dto.RoleDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	role, err := entity.NewRole(cmd.Name, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	res := dto.ToRoleDTO(role)
	return &res, nil
}

func (s *rbacService) GetRole(ctx context.Context, q query.GetRoleByIDQuery) (*dto.RoleDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	res := dto.ToRoleDTO(role)
	return &res, nil
}

func (s *rbacService) ListRoles(ctx context.Context, q query.ListRolesQuery) ([]dto.RoleDTO, pagination.Meta, error) {
	roles, total, err := s.roleRepo.List(ctx, q.Pagination)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	meta := pagination.NewMeta(q.Pagination.Page, q.Pagination.PageSize, total)
	return dto.ToRoleDTOList(roles), meta, nil
}

func (s *rbacService) UpdateRole(ctx context.Context, cmd command.UpdateRoleCommand) (*dto.RoleDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := role.Update(cmd.Name, cmd.Description); err != nil {
		return nil, err
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	res := dto.ToRoleDTO(role)
	return &res, nil
}

func (s *rbacService) DeleteRole(ctx context.Context, cmd command.DeleteRoleCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	role, err := s.roleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return appErrors.NewBusinessError("SYSTEM_ROLE_PROTECTED", fmt.Sprintf("system role '%s' cannot be deleted", role.Name))
	}

	// Fetch affected users for cache invalidation
	affectedUsers, _ := s.roleRepo.GetUserIDsByRoleID(ctx, role.ID)

	if err := s.roleRepo.Delete(ctx, role.ID); err != nil {
		return err
	}

	// Invalidate cache
	if s.permissionCache != nil && len(affectedUsers) > 0 {
		_ = s.permissionCache.InvalidateUsers(ctx, affectedUsers)
	}

	return nil
}

func (s *rbacService) AssignPermissionsToRole(ctx context.Context, cmd command.AssignRolePermissionsCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Verify role exists
	role, err := s.roleRepo.FindByID(ctx, cmd.RoleID)
	if err != nil {
		return err
	}

	// Verify all permissions exist
	for _, pid := range cmd.PermissionIDs {
		if _, err := s.permissionRepo.FindByID(ctx, pid); err != nil {
			return appErrors.NewValidationError(fmt.Sprintf("invalid permission ID: %s", pid))
		}
	}

	if err := s.roleRepo.AssignPermissions(ctx, role.ID, cmd.PermissionIDs); err != nil {
		return err
	}

	// Invalidate cache for all users holding this role
	affectedUsers, _ := s.roleRepo.GetUserIDsByRoleID(ctx, role.ID)
	if s.permissionCache != nil && len(affectedUsers) > 0 {
		_ = s.permissionCache.InvalidateUsers(ctx, affectedUsers)
	}

	return nil
}

func (s *rbacService) GetRolePermissions(ctx context.Context, q query.GetRolePermissionsQuery) (*dto.RoleWithPermissionsDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, q.RoleID)
	if err != nil {
		return nil, err
	}

	perms, err := s.permissionRepo.FindByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}

	return &dto.RoleWithPermissionsDTO{
		Role:        dto.ToRoleDTO(role),
		Permissions: dto.ToPermissionDTOList(perms),
	}, nil
}

// --- Permission Management ---

func (s *rbacService) CreatePermission(ctx context.Context, cmd command.CreatePermissionCommand) (*dto.PermissionDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	perm, err := entity.NewPermission(cmd.Resource, cmd.Action, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := s.permissionRepo.Create(ctx, perm); err != nil {
		return nil, err
	}

	res := dto.ToPermissionDTO(perm)
	return &res, nil
}

func (s *rbacService) GetPermission(ctx context.Context, q query.GetPermissionByIDQuery) (*dto.PermissionDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	perm, err := s.permissionRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	res := dto.ToPermissionDTO(perm)
	return &res, nil
}

func (s *rbacService) ListPermissions(ctx context.Context, q query.ListPermissionsQuery) ([]dto.PermissionDTO, pagination.Meta, error) {
	perms, total, err := s.permissionRepo.List(ctx, q.Pagination)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	meta := pagination.NewMeta(q.Pagination.Page, q.Pagination.PageSize, total)
	return dto.ToPermissionDTOList(perms), meta, nil
}

func (s *rbacService) UpdatePermission(ctx context.Context, cmd command.UpdatePermissionCommand) (*dto.PermissionDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	perm, err := s.permissionRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := perm.Update(cmd.Resource, cmd.Action, cmd.Description); err != nil {
		return nil, err
	}

	if err := s.permissionRepo.Update(ctx, perm); err != nil {
		return nil, err
	}

	// Invalidate affected users cache
	affectedUsers, _ := s.permissionRepo.GetAffectedUserIDsByPermissionID(ctx, perm.ID)
	if s.permissionCache != nil && len(affectedUsers) > 0 {
		_ = s.permissionCache.InvalidateUsers(ctx, affectedUsers)
	}

	res := dto.ToPermissionDTO(perm)
	return &res, nil
}

func (s *rbacService) DeletePermission(ctx context.Context, cmd command.DeletePermissionCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	perm, err := s.permissionRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	affectedUsers, _ := s.permissionRepo.GetAffectedUserIDsByPermissionID(ctx, perm.ID)

	if err := s.permissionRepo.Delete(ctx, perm.ID); err != nil {
		return err
	}

	if s.permissionCache != nil && len(affectedUsers) > 0 {
		_ = s.permissionCache.InvalidateUsers(ctx, affectedUsers)
	}

	return nil
}

// --- User RBAC Management ---

func (s *rbacService) AssignRolesToUser(ctx context.Context, cmd command.AssignUserRolesCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Verify all roles exist
	for _, rid := range cmd.RoleIDs {
		if _, err := s.roleRepo.FindByID(ctx, rid); err != nil {
			return appErrors.NewValidationError(fmt.Sprintf("invalid role ID: %s", rid))
		}
	}

	if err := s.roleRepo.AssignUserRoles(ctx, cmd.UserID, cmd.RoleIDs); err != nil {
		return err
	}

	// Invalidate target user's permission cache
	if s.permissionCache != nil {
		_ = s.permissionCache.InvalidateUser(ctx, cmd.UserID)
	}

	return nil
}

func (s *rbacService) GetUserRoles(ctx context.Context, q query.GetUserRolesQuery) (*dto.UserRolesDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	roles, err := s.roleRepo.GetUserRoles(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.UserRolesDTO{
		UserID: q.UserID,
		Roles:  dto.ToRoleDTOList(roles),
	}, nil
}

func (s *rbacService) GetUserPermissionsDTO(ctx context.Context, q query.GetUserPermissionsQuery) (*dto.UserPermissionsDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	perms, err := s.GetUserPermissions(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.UserPermissionsDTO{
		UserID:      q.UserID,
		Permissions: perms,
	}, nil
}
