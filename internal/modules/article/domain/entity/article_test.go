package entity_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestArticle_NewArticle_Success(t *testing.T) {
	article, err := entity.NewArticle("user-123", "Getting Started with Go", "Full content here.")
	assert.NoError(t, err)
	assert.NotEmpty(t, article.ID)
	assert.Equal(t, "user-123", article.UserID)
	assert.Equal(t, "Getting Started with Go", article.Title)
	assert.Equal(t, "getting-started-with-go", article.Slug)
	assert.Equal(t, entity.ArticleStatusDraft, article.Status)
	assert.False(t, article.CreatedAt.IsZero())
}

func TestArticle_NewArticle_Validation(t *testing.T) {
	t.Run("Empty author user ID", func(t *testing.T) {
		a, err := entity.NewArticle("", "Title", "Content")
		assert.Error(t, err)
		assert.Nil(t, a)
	})

	t.Run("Empty title", func(t *testing.T) {
		a, err := entity.NewArticle("user-1", "", "Content")
		assert.Error(t, err)
		assert.Nil(t, a)
	})

	t.Run("Title too short", func(t *testing.T) {
		a, err := entity.NewArticle("user-1", "ab", "Content")
		assert.Error(t, err)
		assert.Nil(t, a)
	})

	t.Run("Empty content", func(t *testing.T) {
		a, err := entity.NewArticle("user-1", "Valid Title", "")
		assert.Error(t, err)
		assert.Nil(t, a)
	})
}

func TestArticle_Update(t *testing.T) {
	a, err := entity.NewArticle("user-1", "Old Title", "Old Content")
	assert.NoError(t, err)

	err = a.Update("New Title Here", "New Content", entity.ArticleStatusPublished)
	assert.NoError(t, err)
	assert.Equal(t, "New Title Here", a.Title)
	assert.Equal(t, "new-title-here", a.Slug)
	assert.Equal(t, "New Content", a.Content)
	assert.Equal(t, entity.ArticleStatusPublished, a.Status)

	err = a.Update("", "", "invalid_status")
	assert.Error(t, err)
}
