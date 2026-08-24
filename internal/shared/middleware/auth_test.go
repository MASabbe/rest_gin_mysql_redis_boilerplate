package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenSvc := jwtInfra.NewJWTService(config.JWTConfig{
		Secret:          "secret-key-that-is-at-least-32-chars-long!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-app",
		Audience:        "test-users",
	})

	pair, err := tokenSvc.GenerateTokenPair("usr-123", "usr@example.com")
	assert.NoError(t, err)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(tokenSvc))
	router.GET("/protected", func(c *gin.Context) {
		userID, ok := middleware.GetAuthenticatedUserID(c)
		assert.True(t, ok)
		assert.Equal(t, "usr-123", userID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenSvc := jwtInfra.NewJWTService(config.JWTConfig{
		Secret:          "secret-key-that-is-at-least-32-chars-long!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-app",
		Audience:        "test-users",
	})

	router := gin.New()
	router.Use(middleware.AuthMiddleware(tokenSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenSvc := jwtInfra.NewJWTService(config.JWTConfig{
		Secret:          "secret-key-that-is-at-least-32-chars-long!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-app",
		Audience:        "test-users",
	})

	router := gin.New()
	router.Use(middleware.AuthMiddleware(tokenSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic some-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
