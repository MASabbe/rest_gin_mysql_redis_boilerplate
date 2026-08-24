package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/handler"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/request"
	jwtInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/password"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupTestAuthHandlerRouter() (*gin.Engine, *handler.AuthHandler, application.AuthService, *jwtInfra.JWTService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := persistence.NewInMemoryUserRepository()
	tokenRepo := persistence.NewInMemoryTokenRepository()
	pwdHasher := password.NewBcryptHasher(bcrypt.MinCost)
	jwtCfg := config.JWTConfig{
		Secret:          "test-jwt-secret-key-at-least-32-chars-long!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-app",
		Audience:        "test-users",
	}
	tokenSvc := jwtInfra.NewJWTService(jwtCfg)
	authSvc := application.NewAuthService(userRepo, tokenRepo, pwdHasher, tokenSvc, 7*24*time.Hour)
	authHandler := handler.NewAuthHandler(authSvc)

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/refresh", authHandler.Refresh)
		api.POST("/logout", authHandler.Logout)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(tokenSvc))
		{
			protected.GET("/me", authHandler.Me)
		}
	}

	return r, authHandler, authSvc, tokenSvc.(*jwtInfra.JWTService)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	r, _, _, _ := setupTestAuthHandlerRouter()

	reqBody, _ := json.Marshal(request.RegisterRequest{
		Email:    "testuser@example.com",
		Password: "SecurePassword123!",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User registered successfully", resp.Message)
}

func TestAuthHandler_Register_InvalidPayload(t *testing.T) {
	r, _, _, _ := setupTestAuthHandlerRouter()

	reqBody, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_And_Me_Success(t *testing.T) {
	r, _, _, _ := setupTestAuthHandlerRouter()

	// 1. Register
	regBody, _ := json.Marshal(request.RegisterRequest{
		Email:    "flow@example.com",
		Password: "Password123!",
	})
	wReg := httptest.NewRecorder()
	reqReg, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wReg, reqReg)
	assert.Equal(t, http.StatusCreated, wReg.Code)

	// 2. Login
	loginBody, _ := json.Marshal(request.LoginRequest{
		Email:    "flow@example.com",
		Password: "Password123!",
	})
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wLogin, reqLogin)

	assert.Equal(t, http.StatusOK, wLogin.Code)
	var loginResp struct {
		Success bool `json:"success"`
		Data    struct {
			Tokens struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	err := json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.Data.Tokens.AccessToken)

	// 3. Call /me with Bearer token
	wMe := httptest.NewRecorder()
	reqMe, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResp.Data.Tokens.AccessToken)
	r.ServeHTTP(wMe, reqMe)

	assert.Equal(t, http.StatusOK, wMe.Code)

	// 4. Refresh Token
	refreshBody, _ := json.Marshal(request.RefreshTokenRequest{
		RefreshToken: loginResp.Data.Tokens.RefreshToken,
	})
	wRefresh := httptest.NewRecorder()
	reqRefresh, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
	reqRefresh.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wRefresh, reqRefresh)
	assert.Equal(t, http.StatusOK, wRefresh.Code)

	// 5. Logout
	logoutBody, _ := json.Marshal(request.LogoutRequest{
		RefreshToken: loginResp.Data.Tokens.RefreshToken,
	})
	wLogout := httptest.NewRecorder()
	reqLogout, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(logoutBody))
	reqLogout.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wLogout, reqLogout)
	assert.Equal(t, http.StatusOK, wLogout.Code)
}
