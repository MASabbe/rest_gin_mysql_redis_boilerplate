package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authJWT "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	rbacHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, application.RBACService, *persistence.InMemoryRoleRepository, *persistence.InMemoryPermissionRepository, string, string) {
	gin.SetMode(gin.TestMode)

	jwtCfg := config.JWTConfig{
		Secret:          "test-secret-minimum-32-chars-length-required-here",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test-issuer",
		Audience:        "test-aud",
	}
	jwtService := authJWT.NewJWTService(jwtCfg)

	roleRepo := persistence.NewInMemoryRoleRepository()
	permRepo := persistence.NewInMemoryPermissionRepository(roleRepo)
	rbacService := application.NewRBACService(roleRepo, permRepo, nil)

	// Seed seeder
	seeder := persistence.NewSeeder(roleRepo, permRepo)
	_ = seeder.Seed(context.Background())

	// Admin user with admin role
	adminUserID := "admin-user-id"
	adminRole, _ := roleRepo.FindByName(context.Background(), "admin")
	_ = roleRepo.AssignUserRoles(context.Background(), adminUserID, []string{adminRole.ID})
	adminTokens, _ := jwtService.GenerateTokenPair(adminUserID, "admin@example.com")

	// Regular user without admin permissions
	regularUserID := "regular-user-id"
	regularRole, _ := entity.NewRole("viewer", "Viewer")
	_ = roleRepo.Create(context.Background(), regularRole)
	_ = roleRepo.AssignUserRoles(context.Background(), regularUserID, []string{regularRole.ID})
	regularTokens, _ := jwtService.GenerateTokenPair(regularUserID, "user@example.com")

	engine := gin.New()
	apiV1 := engine.Group("/api/v1")
	rbacHTTP.RegisterRoutes(apiV1, jwtService, rbacService)

	return engine, rbacService, roleRepo, permRepo, adminTokens.AccessToken, regularTokens.AccessToken
}

func TestRBACHandlers_RoleCRUD_And_Permissions(t *testing.T) {
	router, _, _, _, adminToken, regularToken := setupTestRouter()

	t.Run("Anonymous request returns 401 Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Regular user without role:read returns 403 Forbidden", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+regularToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Admin with role:read can list roles", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	var createdRoleID string
	t.Run("Admin can create new role", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name":        "moderator",
			"description": "Content moderator",
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]any)
		createdRoleID = data["id"].(string)
		assert.NotEmpty(t, createdRoleID)
	})

	t.Run("Admin can update role", func(t *testing.T) {
		require.NotEmpty(t, createdRoleID)
		body, _ := json.Marshal(map[string]string{
			"name":        "senior-moderator",
			"description": "Senior content moderator",
		})
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+createdRoleID, bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Admin can assign permissions to role", func(t *testing.T) {
		require.NotEmpty(t, createdRoleID)
		body, _ := json.Marshal(map[string]any{
			"permission_ids": []string{},
		})
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+createdRoleID+"/permissions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Admin can delete custom role", func(t *testing.T) {
		require.NotEmpty(t, createdRoleID)
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+createdRoleID, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRBACHandlers_PermissionCRUD_And_UserRoles(t *testing.T) {
	router, _, _, _, adminToken, _ := setupTestRouter()

	var createdPermID string
	t.Run("Admin can create permission", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"resource":    "document",
			"action":      "download",
			"description": "Download document",
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/permissions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]any)
		createdPermID = data["id"].(string)
		assert.NotEmpty(t, createdPermID)
	})

	t.Run("Admin can list permissions with pagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/permissions?page=1&page_size=5", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Admin can assign roles to user", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"role_ids": []string{},
		})
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/target-user-123/roles", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authenticated user can view their own permissions", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me/permissions", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
