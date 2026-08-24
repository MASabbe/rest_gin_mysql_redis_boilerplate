package queryparam_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractSorting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	allowlist := []string{"created_at", "title", "updated_at"}

	t.Run("Default values when missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)

		s := queryparam.ExtractSorting(c, allowlist, "created_at", "DESC")
		assert.Equal(t, "created_at", s.Field)
		assert.Equal(t, "DESC", s.Order)
		assert.Equal(t, "created_at DESC", s.SQLOrderBy())
	})

	t.Run("Valid allowed field and order", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test?sort=title&order=asc", nil)

		s := queryparam.ExtractSorting(c, allowlist, "created_at", "DESC")
		assert.Equal(t, "title", s.Field)
		assert.Equal(t, "ASC", s.Order)
		assert.Equal(t, "title ASC", s.SQLOrderBy())
	})

	t.Run("Disallow dangerous field injection", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test?sort=password_hash;DROP TABLE users--&order=asc", nil)

		s := queryparam.ExtractSorting(c, allowlist, "created_at", "DESC")
		assert.Equal(t, "created_at", s.Field)
		assert.Equal(t, "ASC", s.Order)
	})
}
