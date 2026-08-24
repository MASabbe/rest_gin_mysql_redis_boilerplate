package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	articleApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application"
	articleHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/delivery/http"
	articlePersistence "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/infrastructure/persistence"
	authJWT "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthorizer struct {
	allowedPerms map[string]bool
}

func (m *mockAuthorizer) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	return m.allowedPerms[permission], nil
}

func (m *mockAuthorizer) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}

func setupArticleTestRouter() (*gin.Engine, articleApp.ArticleService, string, string) {
	gin.SetMode(gin.TestMode)

	jwtCfg := config.JWTConfig{
		Secret:          "test-secret-minimum-32-chars-length-required-here",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test-issuer",
		Audience:        "test-aud",
	}
	jwtService := authJWT.NewJWTService(jwtCfg)

	authorizer := &mockAuthorizer{
		allowedPerms: map[string]bool{
			"article:read":   true,
			"article:create": true,
			"article:update": true,
			"article:delete": true,
		},
	}

	articleRepo := articlePersistence.NewInMemoryArticleRepository()
	articleService := articleApp.NewArticleService(articleRepo)

	authorTokens, _ := jwtService.GenerateTokenPair("author-123", "author@example.com")
	strangerTokens, _ := jwtService.GenerateTokenPair("stranger-456", "stranger@example.com")

	engine := gin.New()
	apiV1 := engine.Group("/api/v1")
	articleHTTP.RegisterRoutes(apiV1, jwtService, authorizer, articleService)

	return engine, articleService, authorTokens.AccessToken, strangerTokens.AccessToken
}

func TestArticleHandler_CRUD_And_Endpoints(t *testing.T) {
	router, _, authorToken, _ := setupArticleTestRouter()

	var createdArticleID string
	var createdArticleSlug string

	t.Run("Create Article", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"title":   "Building Resilient Go Backends",
			"content": "Full article guide here.",
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/articles", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authorToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]any)
		createdArticleID = data["id"].(string)
		createdArticleSlug = data["slug"].(string)
		assert.NotEmpty(t, createdArticleID)
		assert.Equal(t, "building-resilient-go-backends", createdArticleSlug)
	})

	t.Run("Get Article By ID", func(t *testing.T) {
		require.NotEmpty(t, createdArticleID)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/"+createdArticleID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get Article By Slug", func(t *testing.T) {
		require.NotEmpty(t, createdArticleSlug)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles/slug/"+createdArticleSlug, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List Articles with Sorting and Pagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/articles?page=1&page_size=10&sort=title&order=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update Article", func(t *testing.T) {
		require.NotEmpty(t, createdArticleID)
		body, _ := json.Marshal(map[string]string{
			"title":  "Building Resilient Go Backends - Rev 2",
			"status": "published",
		})
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/articles/"+createdArticleID, bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authorToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Article", func(t *testing.T) {
		require.NotEmpty(t, createdArticleID)
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/articles/"+createdArticleID, nil)
		req.Header.Set("Authorization", "Bearer "+authorToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
