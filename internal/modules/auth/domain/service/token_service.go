package service

import "time"

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// TokenClaims represents the extracted and validated claims from a JWT.
type TokenClaims struct {
	TokenID   string    `json:"jti"`
	UserID    string    `json:"sub"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
}

// TokenPair represents generated access and refresh tokens with metadata.
type TokenPair struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	TokenType             string `json:"token_type"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`  // seconds
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"` // seconds
	RefreshTokenID        string `json:"-"`
}

// TokenService abstracts generation and validation of authentication tokens.
type TokenService interface {
	GenerateTokenPair(userID, email string) (*TokenPair, error)
	ValidateAccessToken(tokenStr string) (*TokenClaims, error)
	ValidateRefreshToken(tokenStr string) (*TokenClaims, error)
}
