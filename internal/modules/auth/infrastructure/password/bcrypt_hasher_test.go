package password_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/password"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_HashAndCompare_Success(t *testing.T) {
	hasher := password.NewBcryptHasher(bcrypt.MinCost) // use MinCost for fast tests

	pwd := "MySecurePassword123!"
	hashed, err := hasher.Hash(pwd)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, pwd, hashed)

	err = hasher.Compare(hashed, pwd)
	assert.NoError(t, err)
}

func TestBcryptHasher_Compare_Failure(t *testing.T) {
	hasher := password.NewBcryptHasher(bcrypt.MinCost)

	hashed, err := hasher.Hash("CorrectPassword")
	assert.NoError(t, err)

	err = hasher.Compare(hashed, "WrongPassword")
	assert.Error(t, err)
}

func TestBcryptHasher_EmptyPassword(t *testing.T) {
	hasher := password.NewBcryptHasher(bcrypt.MinCost)

	_, err := hasher.Hash("")
	assert.Error(t, err)
}
