package entity_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestUser_NewUser_Success(t *testing.T) {
	u, err := entity.NewUser("test@example.com", "hashed_password")
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.NotEmpty(t, u.ID)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, entity.UserStatusActive, u.Status)
	assert.True(t, u.IsActive())
}

func TestUser_NewUser_InvalidEmail(t *testing.T) {
	_, err := entity.NewUser("", "hashed_password")
	assert.Error(t, err)

	_, err = entity.NewUser("notanemail", "hashed_password")
	assert.Error(t, err)
}

func TestUser_NewUser_EmptyPasswordHash(t *testing.T) {
	_, err := entity.NewUser("valid@example.com", "   ")
	assert.Error(t, err)
}
