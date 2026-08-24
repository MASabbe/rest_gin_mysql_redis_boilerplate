package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	jwtPkg "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type customClaims struct {
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwtPkg.RegisteredClaims
}

type JWTService struct {
	cfg config.JWTConfig
}

// NewJWTService creates a new JWTService instance.
func NewJWTService(cfg config.JWTConfig) service.TokenService {
	return &JWTService{cfg: cfg}
}

// GenerateTokenPair generates both an access token and a refresh token.
func (j *JWTService) GenerateTokenPair(userID, email string) (*service.TokenPair, error) {
	now := time.Now().UTC()

	// 1. Generate Access Token
	accessJTI := uuid.New().String()
	accessExpiry := now.Add(j.cfg.AccessTokenTTL)
	accessClaims := customClaims{
		Email:     email,
		TokenType: service.TokenTypeAccess,
		RegisteredClaims: jwtPkg.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID,
			Issuer:    j.cfg.Issuer,
			Audience:  jwtPkg.ClaimStrings{j.cfg.Audience},
			IssuedAt:  jwtPkg.NewNumericDate(now),
			NotBefore: jwtPkg.NewNumericDate(now),
			ExpiresAt: jwtPkg.NewNumericDate(accessExpiry),
		},
	}

	accessToken := jwtPkg.NewWithClaims(jwtPkg.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(j.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Generate Refresh Token
	refreshJTI := uuid.New().String()
	refreshExpiry := now.Add(j.cfg.RefreshTokenTTL)
	refreshClaims := customClaims{
		Email:     email,
		TokenType: service.TokenTypeRefresh,
		RegisteredClaims: jwtPkg.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   userID,
			Issuer:    j.cfg.Issuer,
			Audience:  jwtPkg.ClaimStrings{j.cfg.Audience},
			IssuedAt:  jwtPkg.NewNumericDate(now),
			NotBefore: jwtPkg.NewNumericDate(now),
			ExpiresAt: jwtPkg.NewNumericDate(refreshExpiry),
		},
	}

	refreshToken := jwtPkg.NewWithClaims(jwtPkg.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &service.TokenPair{
		AccessToken:           signedAccessToken,
		RefreshToken:          signedRefreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresIn:  int64(j.cfg.AccessTokenTTL.Seconds()),
		RefreshTokenExpiresIn: int64(j.cfg.RefreshTokenTTL.Seconds()),
		RefreshTokenID:        refreshJTI,
	}, nil
}

// ValidateAccessToken validates signature, claims, algorithm, and verifies token type is access.
func (j *JWTService) ValidateAccessToken(tokenStr string) (*service.TokenClaims, error) {
	claims, err := j.validateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != service.TokenTypeAccess {
		return nil, appErrors.NewUnauthorizedError("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates signature, claims, algorithm, and verifies token type is refresh.
func (j *JWTService) ValidateRefreshToken(tokenStr string) (*service.TokenClaims, error) {
	claims, err := j.validateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != service.TokenTypeRefresh {
		return nil, appErrors.NewUnauthorizedError("invalid token type: expected refresh token")
	}

	return claims, nil
}

func (j *JWTService) validateToken(tokenStr string) (*service.TokenClaims, error) {
	if tokenStr == "" {
		return nil, appErrors.NewUnauthorizedError("missing token")
	}

	token, err := jwtPkg.ParseWithClaims(tokenStr, &customClaims{}, func(token *jwtPkg.Token) (interface{}, error) {
		// Enforce HS256 algorithm explicitly
		if _, ok := token.Method.(*jwtPkg.SigningMethodHMAC); !ok || token.Method.Alg() != jwtPkg.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.cfg.Secret), nil
	},
		jwtPkg.WithIssuer(j.cfg.Issuer),
		jwtPkg.WithAudience(j.cfg.Audience),
		jwtPkg.WithLeeway(5*time.Second),
	)

	if err != nil {
		if errors.Is(err, jwtPkg.ErrTokenExpired) {
			return nil, appErrors.NewUnauthorizedError("token has expired", err)
		}
		return nil, appErrors.NewUnauthorizedError("invalid token signature or claims", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, appErrors.NewUnauthorizedError("invalid token claims")
	}

	return &service.TokenClaims{
		TokenID:   claims.ID,
		UserID:    claims.Subject,
		Email:     claims.Email,
		TokenType: claims.TokenType,
		Issuer:    claims.Issuer,
		Audience:  j.cfg.Audience,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
