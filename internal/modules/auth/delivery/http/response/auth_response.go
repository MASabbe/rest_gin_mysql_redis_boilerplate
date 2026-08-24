package response

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/dto"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	TokenType             string `json:"token_type"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

func FromAuthResultDTO(result *dto.AuthResultDTO) AuthResponse {
	return AuthResponse{
		User: UserResponse{
			ID:        result.User.ID,
			Email:     result.User.Email,
			Status:    string(result.User.Status),
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		},
		Tokens: TokenResponse{
			AccessToken:           result.Tokens.AccessToken,
			RefreshToken:          result.Tokens.RefreshToken,
			TokenType:             result.Tokens.TokenType,
			AccessTokenExpiresIn:  result.Tokens.AccessTokenExpiresIn,
			RefreshTokenExpiresIn: result.Tokens.RefreshTokenExpiresIn,
		},
	}
}

func FromTokenPairDTO(tokens *dto.TokenPairDTO) TokenResponse {
	return TokenResponse{
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		TokenType:             tokens.TokenType,
		AccessTokenExpiresIn:  tokens.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: tokens.RefreshTokenExpiresIn,
	}
}

func FromUserDTO(u *dto.UserDTO) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
