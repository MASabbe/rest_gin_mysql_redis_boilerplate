package entity_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestRole_NewRole_Success(t *testing.T) {
	role, err := entity.NewRole("editor", "Can edit content")
	assert.NoError(t, err)
	assert.NotEmpty(t, role.ID)
	assert.Equal(t, "editor", role.Name)
	assert.Equal(t, "Can edit content", role.Description)
	assert.False(t, role.IsSystem)
	assert.False(t, role.CreatedAt.IsZero())
}

func TestRole_NewSystemRole_Success(t *testing.T) {
	role, err := entity.NewSystemRole("admin", "Administrator")
	assert.NoError(t, err)
	assert.True(t, role.IsSystem)
	assert.Equal(t, "admin", role.Name)
}

func TestRole_NewRole_Validation(t *testing.T) {
	t.Run("Empty name", func(t *testing.T) {
		role, err := entity.NewRole("", "Desc")
		assert.Error(t, err)
		assert.Nil(t, role)
	})

	t.Run("Name too short", func(t *testing.T) {
		role, err := entity.NewRole("a", "Desc")
		assert.Error(t, err)
		assert.Nil(t, role)
	})
}

func TestRole_Update(t *testing.T) {
	role, err := entity.NewRole("editor", "Old desc")
	assert.NoError(t, err)

	err = role.Update("senior-editor", "New desc")
	assert.NoError(t, err)
	assert.Equal(t, "senior-editor", role.Name)
	assert.Equal(t, "New desc", role.Description)

	err = role.Update("", "Empty")
	assert.Error(t, err)
}
