package persistence_test

import (
	"context"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeeder_IdempotentExecution(t *testing.T) {
	roleRepo := persistence.NewInMemoryRoleRepository()
	permRepo := persistence.NewInMemoryPermissionRepository(roleRepo)
	seeder := persistence.NewSeeder(roleRepo, permRepo)
	ctx := context.Background()

	// First execution
	err := seeder.Seed(ctx)
	require.NoError(t, err)

	adminRole, err := roleRepo.FindByName(ctx, "admin")
	require.NoError(t, err)
	assert.True(t, adminRole.IsSystem)

	adminPerms, err := permRepo.FindByRoleID(ctx, adminRole.ID)
	require.NoError(t, err)
	assert.Len(t, adminPerms, len(persistence.CorePermissions))

	// Second execution (must not fail or duplicate)
	err = seeder.Seed(ctx)
	require.NoError(t, err)

	adminPermsSecond, err := permRepo.FindByRoleID(ctx, adminRole.ID)
	require.NoError(t, err)
	assert.Len(t, adminPermsSecond, len(persistence.CorePermissions))
}
