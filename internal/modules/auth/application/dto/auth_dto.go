package dto

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
)

// UserDTO represents a sanitized user representation for API responses.
type UserDTO struct {
	ID             string            `json:"id"`
	Email          string            `json:"email"`
	Status         entity.UserStatus `json:"status"`
	LastLoginAt    *time.Time        `json:"last_login_at,omitempty"`
	LastActivityAt *time.Time        `json:"last_activity_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func ToUserDTO(u *entity.User) UserDTO {
	return UserDTO{
		ID:             u.ID,
		Email:          u.Email,
		Status:         u.Status,
		LastLoginAt:    u.LastLoginAt,
		LastActivityAt: u.LastActivityAt,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

// TokenPairDTO represents the auth token response.
type TokenPairDTO struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	TokenType             string `json:"token_type"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

// AuthResultDTO combines user profile and issued tokens.
type AuthResultDTO struct {
	User   UserDTO      `json:"user"`
	Tokens TokenPairDTO `json:"tokens"`
}
