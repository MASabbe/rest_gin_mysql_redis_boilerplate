package jwt_test

import (
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	jwtInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	jwtPkg "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func testJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:          "very-secure-jwt-secret-key-at-least-32-chars!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
		Audience:        "test-audience",
	}
}

func TestJWTService_GenerateAndValidate_Success(t *testing.T) {
	svc := jwtInfra.NewJWTService(testJWTConfig())

	pair, err := svc.GenerateTokenPair("user-123", "user@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEmpty(t, pair.RefreshTokenID)
	assert.Equal(t, "Bearer", pair.TokenType)

	// Validate Access Token
	accessClaims, err := svc.ValidateAccessToken(pair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", accessClaims.UserID)
	assert.Equal(t, "user@example.com", accessClaims.Email)
	assert.Equal(t, service.TokenTypeAccess, accessClaims.TokenType)
	assert.Equal(t, "test-issuer", accessClaims.Issuer)

	// Validate Refresh Token
	refreshClaims, err := svc.ValidateRefreshToken(pair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", refreshClaims.UserID)
	assert.Equal(t, "user@example.com", refreshClaims.Email)
	assert.Equal(t, service.TokenTypeRefresh, refreshClaims.TokenType)
}

func TestJWTService_TokenTypeMismatch(t *testing.T) {
	svc := jwtInfra.NewJWTService(testJWTConfig())

	pair, err := svc.GenerateTokenPair("user-123", "user@example.com")
	assert.NoError(t, err)

	// Attempt to use Refresh Token as Access Token
	_, err = svc.ValidateAccessToken(pair.RefreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type: expected access token")

	// Attempt to use Access Token as Refresh Token
	_, err = svc.ValidateRefreshToken(pair.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type: expected refresh token")
}

func TestJWTService_ExpiredToken(t *testing.T) {
	cfg := testJWTConfig()
	cfg.AccessTokenTTL = -1 * time.Minute // expired
	cfg.RefreshTokenTTL = -1 * time.Minute
	svc := jwtInfra.NewJWTService(cfg)

	pair, err := svc.GenerateTokenPair("user-123", "user@example.com")
	assert.NoError(t, err)

	_, err = svc.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has expired")

	_, err = svc.ValidateRefreshToken(pair.RefreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has expired")
}

func TestJWTService_InvalidSignature(t *testing.T) {
	svc1 := jwtInfra.NewJWTService(testJWTConfig())

	cfg2 := testJWTConfig()
	cfg2.Secret = "different-secret-key-which-is-at-least-32-chars"
	svc2 := jwtInfra.NewJWTService(cfg2)

	pair, err := svc1.GenerateTokenPair("user-123", "user@example.com")
	assert.NoError(t, err)

	_, err = svc2.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token signature")
}

func TestJWTService_RejectNoneAlgorithm(t *testing.T) {
	// Craft an unsecure token with "none" algorithm
	token := jwtPkg.NewWithClaims(jwtPkg.SigningMethodNone, jwtPkg.MapClaims{
		"sub":        "user-123",
		"email":      "user@example.com",
		"token_type": "access",
		"iss":        "test-issuer",
		"aud":        "test-audience",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString(jwtPkg.UnsafeAllowNoneSignatureType)

	svc := jwtInfra.NewJWTService(testJWTConfig())
	_, err := svc.ValidateAccessToken(tokenStr)
	assert.Error(t, err)
}

func TestJWTService_InvalidClaimsAndMalformed(t *testing.T) {
	svc := jwtInfra.NewJWTService(testJWTConfig())

	// 1. Malformed token string
	_, err := svc.ValidateAccessToken("not.a.valid.jwt.token")
	assert.Error(t, err)

	// 2. Wrong Issuer
	wrongIssToken := jwtPkg.NewWithClaims(jwtPkg.SigningMethodHS256, jwtPkg.MapClaims{
		"sub":        "user-123",
		"email":      "user@example.com",
		"token_type": "access",
		"iss":        "wrong-issuer",
		"aud":        "test-audience",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	wrongIssStr, _ := wrongIssToken.SignedString([]byte("very-secure-jwt-secret-key-at-least-32-chars!"))
	_, err = svc.ValidateAccessToken(wrongIssStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has invalid issuer")

	// 3. Wrong Audience
	wrongAudToken := jwtPkg.NewWithClaims(jwtPkg.SigningMethodHS256, jwtPkg.MapClaims{
		"sub":        "user-123",
		"email":      "user@example.com",
		"token_type": "access",
		"iss":        "test-issuer",
		"aud":        "wrong-audience",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	wrongAudStr, _ := wrongAudToken.SignedString([]byte("very-secure-jwt-secret-key-at-least-32-chars!"))
	_, err = svc.ValidateAccessToken(wrongAudStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token has invalid audience")
}
