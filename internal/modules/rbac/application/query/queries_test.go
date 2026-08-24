package query_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/query"
	"github.com/stretchr/testify/assert"
)

func TestQueries_Validation(t *testing.T) {
	t.Run("GetRoleByIDQuery", func(t *testing.T) {
		assert.NoError(t, query.GetRoleByIDQuery{ID: "r1"}.Validate())
		assert.Error(t, query.GetRoleByIDQuery{ID: ""}.Validate())
	})

	t.Run("GetRolePermissionsQuery", func(t *testing.T) {
		assert.NoError(t, query.GetRolePermissionsQuery{RoleID: "r1"}.Validate())
		assert.Error(t, query.GetRolePermissionsQuery{RoleID: ""}.Validate())
	})

	t.Run("GetPermissionByIDQuery", func(t *testing.T) {
		assert.NoError(t, query.GetPermissionByIDQuery{ID: "p1"}.Validate())
		assert.Error(t, query.GetPermissionByIDQuery{ID: ""}.Validate())
	})

	t.Run("GetUserRolesQuery", func(t *testing.T) {
		assert.NoError(t, query.GetUserRolesQuery{UserID: "u1"}.Validate())
		assert.Error(t, query.GetUserRolesQuery{UserID: ""}.Validate())
	})

	t.Run("GetUserPermissionsQuery", func(t *testing.T) {
		assert.NoError(t, query.GetUserPermissionsQuery{UserID: "u1"}.Validate())
		assert.Error(t, query.GetUserPermissionsQuery{UserID: ""}.Validate())
	})
}
