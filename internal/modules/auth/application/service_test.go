package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/query"
	jwtInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/password"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupTestAuthService() application.AuthService {
	userRepo := persistence.NewInMemoryUserRepository()
	tokenRepo := persistence.NewInMemoryTokenRepository()
	pwdHasher := password.NewBcryptHasher(bcrypt.MinCost)

	jwtCfg := config.JWTConfig{
		Secret:          "test-jwt-secret-key-at-least-32-characters!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-app",
		Audience:        "test-users",
	}
	tokenSvc := jwtInfra.NewJWTService(jwtCfg)

	return application.NewAuthService(userRepo, tokenRepo, pwdHasher, tokenSvc, 7*24*time.Hour)
}

func TestAuthService_Register_Success(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	res, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "test@example.com", res.User.Email)
	assert.NotEmpty(t, res.Tokens.AccessToken)
	assert.NotEmpty(t, res.Tokens.RefreshToken)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "duplicate@example.com",
		Password: "Password123!",
	})
	assert.NoError(t, err)

	_, err = svc.Register(ctx, command.RegisterCommand{
		Email:    "DUPLICATE@example.com", // case-insensitive duplicate check
		Password: "Password123!",
	})
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeConflict, appErr.Type)
}

func TestAuthService_Register_ValidationFailure(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "invalid-email",
		Password: "short",
	})
	assert.Error(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "login@example.com",
		Password: "CorrectPassword123!",
	})
	assert.NoError(t, err)

	res, err := svc.Login(ctx, command.LoginCommand{
		Email:    "login@example.com",
		Password: "CorrectPassword123!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "login@example.com", res.User.Email)
	assert.NotEmpty(t, res.Tokens.AccessToken)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "wrongpass@example.com",
		Password: "CorrectPassword123!",
	})
	assert.NoError(t, err)

	_, err = svc.Login(ctx, command.LoginCommand{
		Email:    "wrongpass@example.com",
		Password: "WrongPassword!",
	})
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeUnauthorized, appErr.Type)
}

func TestAuthService_RefreshToken_Rotation_Success(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	regRes, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "refresh@example.com",
		Password: "Password123!",
	})
	assert.NoError(t, err)

	refreshed, err := svc.RefreshToken(ctx, command.RefreshTokenCommand{
		RefreshToken: regRes.Tokens.RefreshToken,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshed.AccessToken)
	assert.NotEmpty(t, refreshed.RefreshToken)
	assert.NotEqual(t, regRes.Tokens.RefreshToken, refreshed.RefreshToken)

	// Old refresh token must now be revoked (Token Rotation)
	_, err = svc.RefreshToken(ctx, command.RefreshTokenCommand{
		RefreshToken: regRes.Tokens.RefreshToken,
	})
	assert.Error(t, err)
}

func TestAuthService_Logout(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	regRes, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "logout@example.com",
		Password: "Password123!",
	})
	assert.NoError(t, err)

	err = svc.Logout(ctx, command.LogoutCommand{
		RefreshToken: regRes.Tokens.RefreshToken,
	})
	assert.NoError(t, err)

	// Attempt to use refresh token after logout should fail
	_, err = svc.RefreshToken(ctx, command.RefreshTokenCommand{
		RefreshToken: regRes.Tokens.RefreshToken,
	})
	assert.Error(t, err)
}

func TestAuthService_GetMe(t *testing.T) {
	svc := setupTestAuthService()
	ctx := context.Background()

	regRes, err := svc.Register(ctx, command.RegisterCommand{
		Email:    "me@example.com",
		Password: "Password123!",
	})
	assert.NoError(t, err)

	me, err := svc.GetMe(ctx, query.GetUserByIDQuery{
		UserID: regRes.User.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, regRes.User.ID, me.ID)
	assert.Equal(t, "me@example.com", me.Email)
}
