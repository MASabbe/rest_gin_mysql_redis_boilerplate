package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPagination_OffsetAndLimit(t *testing.T) {
	p := pagination.Pagination{Page: 2, PageSize: 15}
	assert.Equal(t, 15, p.Offset())
	assert.Equal(t, 15, p.Limit())

	pFirst := pagination.Pagination{Page: 1, PageSize: 20}
	assert.Equal(t, 0, pFirst.Offset())
	assert.Equal(t, 20, pFirst.Limit())
}

func TestNewMeta(t *testing.T) {
	meta := pagination.NewMeta(1, 20, 45)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.PageSize)
	assert.Equal(t, int64(45), meta.TotalCount)
	assert.Equal(t, 3, meta.TotalPages)

	emptyMeta := pagination.NewMeta(1, 20, 0)
	assert.Equal(t, 0, emptyMeta.TotalPages)
}

func TestExtract_GinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Default values when missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test", nil)

		p := pagination.Extract(c)
		assert.Equal(t, pagination.DefaultPage, p.Page)
		assert.Equal(t, pagination.DefaultPageSize, p.PageSize)
	})

	t.Run("Valid query parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test?page=3&page_size=10", nil)

		p := pagination.Extract(c)
		assert.Equal(t, 3, p.Page)
		assert.Equal(t, 10, p.PageSize)
	})

	t.Run("Enforce maximum page size", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test?page=1&page_size=500", nil)

		p := pagination.Extract(c)
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, pagination.MaxPageSize, p.PageSize)
	})

	t.Run("Ignore negative or malformed parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test?page=-2&page_size=abc", nil)

		p := pagination.Extract(c)
		assert.Equal(t, pagination.DefaultPage, p.Page)
		assert.Equal(t, pagination.DefaultPageSize, p.PageSize)
	})
}
