package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMemoryCache struct {
	data             map[string][]string
	invalidatedUsers []string
}

func newMockMemoryCache() *mockMemoryCache {
	return &mockMemoryCache{
		data:             make(map[string][]string),
		invalidatedUsers: make([]string, 0),
	}
}

func (m *mockMemoryCache) Get(ctx context.Context, userID string) ([]string, bool, error) {
	if val, ok := m.data[userID]; ok {
		return val, true, nil
	}
	return nil, false, nil
}

func (m *mockMemoryCache) Set(ctx context.Context, userID string, permissions []string, ttl time.Duration) error {
	m.data[userID] = permissions
	return nil
}

func (m *mockMemoryCache) InvalidateUser(ctx context.Context, userID string) error {
	delete(m.data, userID)
	m.invalidatedUsers = append(m.invalidatedUsers, userID)
	return nil
}

func (m *mockMemoryCache) InvalidateUsers(ctx context.Context, userIDs []string) error {
	for _, uid := range userIDs {
		delete(m.data, uid)
		m.invalidatedUsers = append(m.invalidatedUsers, uid)
	}
	return nil
}

func setupRBACService() (application.RBACService, *persistence.InMemoryRoleRepository, *persistence.InMemoryPermissionRepository, *mockMemoryCache) {
	roleRepo := persistence.NewInMemoryRoleRepository()
	permRepo := persistence.NewInMemoryPermissionRepository(roleRepo)
	cacheMock := newMockMemoryCache()
	svc := application.NewRBACService(roleRepo, permRepo, cacheMock, time.Hour)
	return svc, roleRepo, permRepo, cacheMock
}

func TestRBACService_RoleManagement(t *testing.T) {
	svc, _, _, _ := setupRBACService()
	ctx := context.Background()

	// 1. Create Role
	roleDTO, err := svc.CreateRole(ctx, command.CreateRoleCommand{
		Name:        "manager",
		Description: "Store manager",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, roleDTO.ID)
	assert.Equal(t, "manager", roleDTO.Name)

	// 2. Duplicate Role Name Conflict
	_, err = svc.CreateRole(ctx, command.CreateRoleCommand{
		Name:        "manager",
		Description: "Duplicate",
	})
	assert.Error(t, err)

	// 3. Get Role
	fetched, err := svc.GetRole(ctx, query.GetRoleByIDQuery{ID: roleDTO.ID})
	require.NoError(t, err)
	assert.Equal(t, "manager", fetched.Name)

	// 4. List Roles
	list, meta, err := svc.ListRoles(ctx, query.ListRolesQuery{
		Pagination: pagination.Pagination{Page: 1, PageSize: 10},
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), meta.TotalCount)

	// 5. Update Role
	updated, err := svc.UpdateRole(ctx, command.UpdateRoleCommand{
		ID:          roleDTO.ID,
		Name:        "senior-manager",
		Description: "Updated desc",
	})
	require.NoError(t, err)
	assert.Equal(t, "senior-manager", updated.Name)

	// 6. Delete Role
	err = svc.DeleteRole(ctx, command.DeleteRoleCommand{ID: roleDTO.ID})
	require.NoError(t, err)

	_, err = svc.GetRole(ctx, query.GetRoleByIDQuery{ID: roleDTO.ID})
	assert.Error(t, err)
}

func TestRBACService_SystemRoleProtection(t *testing.T) {
	svc, roleRepo, _, _ := setupRBACService()
	ctx := context.Background()

	adminRole, err := entity.NewSystemRole("admin", "Administrator")
	require.NoError(t, err)
	require.NoError(t, roleRepo.Create(ctx, adminRole))

	// Attempting to delete system role must be blocked
	err = svc.DeleteRole(ctx, command.DeleteRoleCommand{ID: adminRole.ID})
	assert.Error(t, err)
}

func TestRBACService_PermissionManagement(t *testing.T) {
	svc, _, _, _ := setupRBACService()
	ctx := context.Background()

	// 1. Create Permission
	permDTO, err := svc.CreatePermission(ctx, command.CreatePermissionCommand{
		Resource:    "user",
		Action:      "delete",
		Description: "Delete user accounts",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, permDTO.ID)
	assert.Equal(t, "user:delete", permDTO.Name)

	// 2. Get Permission
	fetched, err := svc.GetPermission(ctx, query.GetPermissionByIDQuery{ID: permDTO.ID})
	require.NoError(t, err)
	assert.Equal(t, "user:delete", fetched.Name)

	// 3. List Permissions
	list, meta, err := svc.ListPermissions(ctx, query.ListPermissionsQuery{
		Pagination: pagination.Pagination{Page: 1, PageSize: 10},
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), meta.TotalCount)

	// 4. Update Permission
	updated, err := svc.UpdatePermission(ctx, command.UpdatePermissionCommand{
		ID:          permDTO.ID,
		Resource:    "user",
		Action:      "archive",
		Description: "Archive user",
	})
	require.NoError(t, err)
	assert.Equal(t, "user:archive", updated.Name)

	// 5. Delete Permission
	err = svc.DeletePermission(ctx, command.DeletePermissionCommand{ID: permDTO.ID})
	require.NoError(t, err)
}

func TestRBACService_AuthorizationFlow_And_Cache(t *testing.T) {
	svc, roleRepo, permRepo, cacheMock := setupRBACService()
	ctx := context.Background()

	// 1. Create Role & Permissions
	role, _ := entity.NewRole("editor", "Editor")
	_ = roleRepo.Create(ctx, role)

	perm1, _ := entity.NewPermission("article", "create", "Create article")
	perm2, _ := entity.NewPermission("article", "publish", "Publish article")
	_ = permRepo.Create(ctx, perm1)
	_ = permRepo.Create(ctx, perm2)

	// 2. Assign permissions to role
	err := svc.AssignPermissionsToRole(ctx, command.AssignRolePermissionsCommand{
		RoleID:        role.ID,
		PermissionIDs: []string{perm1.ID, perm2.ID},
	})
	require.NoError(t, err)

	// 3. Assign role to user
	userID := "user-uuid-101"
	err = svc.AssignRolesToUser(ctx, command.AssignUserRolesCommand{
		UserID:  userID,
		RoleIDs: []string{role.ID},
	})
	require.NoError(t, err)

	// 4. Check HasPermission (Cache Miss -> DB -> Cache Populated)
	hasPerm, err := svc.HasPermission(ctx, userID, "article:create")
	require.NoError(t, err)
	assert.True(t, hasPerm)

	hasPerm, err = svc.HasPermission(ctx, userID, "user:delete")
	require.NoError(t, err)
	assert.False(t, hasPerm)

	// Verify Cache was populated
	cachedPerms, hit, _ := cacheMock.Get(ctx, userID)
	assert.True(t, hit)
	assert.Contains(t, cachedPerms, "article:create")
	assert.Contains(t, cachedPerms, "article:publish")

	// 5. Check Cache Invalidation on Role Permission update
	err = svc.AssignPermissionsToRole(ctx, command.AssignRolePermissionsCommand{
		RoleID:        role.ID,
		PermissionIDs: []string{perm1.ID}, // Removed perm2
	})
	require.NoError(t, err)

	// Cache should be evicted for affected user
	_, hit, _ = cacheMock.Get(ctx, userID)
	assert.False(t, hit)

	// Next authorization query repopulates with updated permissions
	hasPerm, err = svc.HasPermission(ctx, userID, "article:publish")
	require.NoError(t, err)
	assert.False(t, hasPerm) // No longer has permission
}

func TestRBACService_GetUserRoles_And_PermissionsDTO(t *testing.T) {
	svc, roleRepo, permRepo, _ := setupRBACService()
	ctx := context.Background()

	role, _ := entity.NewRole("viewer", "Viewer")
	_ = roleRepo.Create(ctx, role)
	perm, _ := entity.NewPermission("dashboard", "view", "View dashboard")
	_ = permRepo.Create(ctx, perm)
	_ = roleRepo.AssignPermissions(ctx, role.ID, []string{perm.ID})

	userID := "user-123"
	_ = roleRepo.AssignUserRoles(ctx, userID, []string{role.ID})

	// GetUserRoles
	userRoles, err := svc.GetUserRoles(ctx, query.GetUserRolesQuery{UserID: userID})
	require.NoError(t, err)
	assert.Equal(t, userID, userRoles.UserID)
	assert.Len(t, userRoles.Roles, 1)

	// GetUserPermissionsDTO
	userPerms, err := svc.GetUserPermissionsDTO(ctx, query.GetUserPermissionsQuery{UserID: userID})
	require.NoError(t, err)
	assert.Equal(t, userID, userPerms.UserID)
	assert.Contains(t, userPerms.Permissions, "dashboard:view")

	// GetRolePermissions
	rolePerms, err := svc.GetRolePermissions(ctx, query.GetRolePermissionsQuery{RoleID: role.ID})
	require.NoError(t, err)
	assert.Equal(t, role.ID, rolePerms.Role.ID)
	assert.Len(t, rolePerms.Permissions, 1)
}
