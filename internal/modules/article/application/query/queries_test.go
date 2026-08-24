package query_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/query"
	"github.com/stretchr/testify/assert"
)

func TestArticleQueries_Validation(t *testing.T) {
	t.Run("GetArticleByIDQuery", func(t *testing.T) {
		assert.NoError(t, query.GetArticleByIDQuery{ID: "a1"}.Validate())
		assert.Error(t, query.GetArticleByIDQuery{ID: ""}.Validate())
	})

	t.Run("GetArticleBySlugQuery", func(t *testing.T) {
		assert.NoError(t, query.GetArticleBySlugQuery{Slug: "article-slug"}.Validate())
		assert.Error(t, query.GetArticleBySlugQuery{Slug: ""}.Validate())
	})
}
