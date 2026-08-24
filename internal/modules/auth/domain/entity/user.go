package entity

import (
	"net/mail"
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents the User domain entity.
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose in JSON serialization
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NewUser creates and validates a new User entity.
func NewUser(email, passwordHash string) (*User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if err := ValidateEmail(cleanEmail); err != nil {
		return nil, err
	}

	if strings.TrimSpace(passwordHash) == "" {
		return nil, appErrors.NewValidationError("password hash cannot be empty")
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New().String(),
		Email:        cleanEmail,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidateEmail validates email format.
func ValidateEmail(email string) error {
	if email == "" {
		return appErrors.NewValidationError("email is required", appErrors.FieldError{Field: "email", Message: "cannot be empty"})
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return appErrors.NewValidationError("invalid email format", appErrors.FieldError{Field: "email", Message: "invalid email format"})
	}
	return nil
}

// IsActive checks if the user account is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
