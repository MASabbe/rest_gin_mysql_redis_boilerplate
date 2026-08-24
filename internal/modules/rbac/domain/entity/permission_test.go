package entity_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestPermission_NewPermission_Success(t *testing.T) {
	perm, err := entity.NewPermission("user", "read", "Can read user profiles")
	assert.NoError(t, err)
	assert.NotEmpty(t, perm.ID)
	assert.Equal(t, "user:read", perm.Name)
	assert.Equal(t, "user", perm.Resource)
	assert.Equal(t, "read", perm.Action)
	assert.Equal(t, "Can read user profiles", perm.Description)
	assert.False(t, perm.CreatedAt.IsZero())
}

func TestPermission_NewPermission_Validation(t *testing.T) {
	t.Run("Empty resource", func(t *testing.T) {
		perm, err := entity.NewPermission("", "read", "Desc")
		assert.Error(t, err)
		assert.Nil(t, perm)
	})

	t.Run("Empty action", func(t *testing.T) {
		perm, err := entity.NewPermission("user", "", "Desc")
		assert.Error(t, err)
		assert.Nil(t, perm)
	})
}

func TestPermission_Update(t *testing.T) {
	perm, err := entity.NewPermission("user", "read", "Old desc")
	assert.NoError(t, err)

	err = perm.Update("user", "list", "New desc")
	assert.NoError(t, err)
	assert.Equal(t, "user:list", perm.Name)
	assert.Equal(t, "user", perm.Resource)
	assert.Equal(t, "list", perm.Action)

	err = perm.Update("", "list", "Empty")
	assert.Error(t, err)
}
