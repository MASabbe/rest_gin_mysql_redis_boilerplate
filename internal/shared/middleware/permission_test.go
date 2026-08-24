package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockAuthzService struct {
	hasPermissionFunc func(ctx context.Context, userID, permission string) (bool, error)
}

func (m *mockAuthzService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	if m.hasPermissionFunc != nil {
		return m.hasPermissionFunc(ctx, userID, permission)
	}
	return false, nil
}

func (m *mockAuthzService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing user_id in context returns 401 Unauthorized", func(t *testing.T) {
		mockService := &mockAuthzService{}
		r := gin.New()
		r.GET("/protected", middleware.RequirePermission(mockService, "user:read"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("User lacks permission returns 403 Forbidden", func(t *testing.T) {
		mockService := &mockAuthzService{
			hasPermissionFunc: func(ctx context.Context, userID, permission string) (bool, error) {
				return false, nil
			},
		}

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextUserIDKey, "user-123")
			c.Next()
		})
		r.GET("/protected", middleware.RequirePermission(mockService, "user:read"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("User possesses permission returns 200 OK", func(t *testing.T) {
		mockService := &mockAuthzService{
			hasPermissionFunc: func(ctx context.Context, userID, permission string) (bool, error) {
				if userID == "user-123" && permission == "user:read" {
					return true, nil
				}
				return false, nil
			},
		}

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextUserIDKey, "user-123")
			c.Next()
		})
		r.GET("/protected", middleware.RequirePermission(mockService, "user:read"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authz service error returns 500 Internal Server Error", func(t *testing.T) {
		mockService := &mockAuthzService{
			hasPermissionFunc: func(ctx context.Context, userID, permission string) (bool, error) {
				return false, errors.New("database failure")
			},
		}

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextUserIDKey, "user-123")
			c.Next()
		})
		r.GET("/protected", middleware.RequirePermission(mockService, "user:read"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
