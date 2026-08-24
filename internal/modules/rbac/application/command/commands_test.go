package command_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application/command"
	"github.com/stretchr/testify/assert"
)

func TestRoleCommands_Validation(t *testing.T) {
	t.Run("CreateRoleCommand", func(t *testing.T) {
		assert.NoError(t, command.CreateRoleCommand{Name: "editor"}.Validate())
		assert.Error(t, command.CreateRoleCommand{Name: ""}.Validate())
		assert.Error(t, command.CreateRoleCommand{Name: "a"}.Validate())
	})

	t.Run("UpdateRoleCommand", func(t *testing.T) {
		assert.NoError(t, command.UpdateRoleCommand{ID: "r1", Name: "editor"}.Validate())
		assert.Error(t, command.UpdateRoleCommand{ID: "", Name: "editor"}.Validate())
		assert.Error(t, command.UpdateRoleCommand{ID: "r1", Name: ""}.Validate())
	})

	t.Run("DeleteRoleCommand", func(t *testing.T) {
		assert.NoError(t, command.DeleteRoleCommand{ID: "r1"}.Validate())
		assert.Error(t, command.DeleteRoleCommand{ID: ""}.Validate())
	})

	t.Run("AssignRolePermissionsCommand", func(t *testing.T) {
		assert.NoError(t, command.AssignRolePermissionsCommand{RoleID: "r1"}.Validate())
		assert.Error(t, command.AssignRolePermissionsCommand{RoleID: ""}.Validate())
	})

	t.Run("AssignUserRolesCommand", func(t *testing.T) {
		assert.NoError(t, command.AssignUserRolesCommand{UserID: "u1"}.Validate())
		assert.Error(t, command.AssignUserRolesCommand{UserID: ""}.Validate())
	})
}

func TestPermissionCommands_Validation(t *testing.T) {
	t.Run("CreatePermissionCommand", func(t *testing.T) {
		assert.NoError(t, command.CreatePermissionCommand{Resource: "user", Action: "read"}.Validate())
		assert.Error(t, command.CreatePermissionCommand{Resource: "", Action: "read"}.Validate())
		assert.Error(t, command.CreatePermissionCommand{Resource: "user", Action: ""}.Validate())
	})

	t.Run("UpdatePermissionCommand", func(t *testing.T) {
		assert.NoError(t, command.UpdatePermissionCommand{ID: "p1", Resource: "user", Action: "read"}.Validate())
		assert.Error(t, command.UpdatePermissionCommand{ID: "", Resource: "user", Action: "read"}.Validate())
		assert.Error(t, command.UpdatePermissionCommand{ID: "p1", Resource: "", Action: "read"}.Validate())
		assert.Error(t, command.UpdatePermissionCommand{ID: "p1", Resource: "user", Action: ""}.Validate())
	})

	t.Run("DeletePermissionCommand", func(t *testing.T) {
		assert.NoError(t, command.DeletePermissionCommand{ID: "p1"}.Validate())
		assert.Error(t, command.DeletePermissionCommand{ID: ""}.Validate())
	})
}
