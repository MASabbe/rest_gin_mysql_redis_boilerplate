package password

import (
	"fmt"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the specified cost (default 12).
func NewBcryptHasher(cost ...int) service.PasswordHasher {
	c := bcrypt.DefaultCost // 10
	if len(cost) > 0 && cost[0] >= bcrypt.MinCost && cost[0] <= bcrypt.MaxCost {
		c = cost[0]
	} else {
		c = 12 // Secure production default
	}
	return &BcryptHasher{cost: c}
}

// Hash securely hashes a plain text password.
func (b *BcryptHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", appErrors.NewValidationError("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// Compare verifies a hashed password against a plain password.
func (b *BcryptHasher) Compare(hashedPassword, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return appErrors.NewUnauthorizedError("invalid email or password")
	}
	return nil
}
