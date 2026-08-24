package application_test

import (
	"context"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticleService_CRUD_And_Ownership(t *testing.T) {
	repo := persistence.NewInMemoryArticleRepository()
	svc := application.NewArticleService(repo)
	ctx := context.Background()

	authorID := "author-123"
	strangerID := "stranger-456"

	// 1. Create Article
	articleDTO, err := svc.CreateArticle(ctx, command.CreateArticleCommand{
		UserID:  authorID,
		Title:   "Clean Architecture in Go",
		Content: "Writing clean decoupled software.",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, articleDTO.ID)
	assert.Equal(t, "clean-architecture-in-go", articleDTO.Slug)
	assert.Equal(t, string(entity.ArticleStatusDraft), articleDTO.Status)

	// 2. Get Article by ID and Slug
	fetched, err := svc.GetArticle(ctx, query.GetArticleByIDQuery{ID: articleDTO.ID})
	require.NoError(t, err)
	assert.Equal(t, articleDTO.Title, fetched.Title)

	bySlug, err := svc.GetArticleBySlug(ctx, query.GetArticleBySlugQuery{Slug: articleDTO.Slug})
	require.NoError(t, err)
	assert.Equal(t, articleDTO.ID, bySlug.ID)

	// 3. List Articles with Filtering
	list, meta, err := svc.ListArticles(ctx, query.ListArticlesQuery{
		UserID:     authorID,
		Status:     entity.ArticleStatusDraft,
		Pagination: pagination.Pagination{Page: 1, PageSize: 10},
		Sorting:    queryparam.Sorting{Field: "created_at", Order: "DESC"},
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), meta.TotalCount)

	// 4. Update Article - Ownership Guard
	_, err = svc.UpdateArticle(ctx, command.UpdateArticleCommand{
		ID:      articleDTO.ID,
		UserID:  strangerID, // Stranger should not be allowed
		Title:   "Hacked Title",
		Content: "Hacked Content",
		Status:  entity.ArticleStatusPublished,
	})
	assert.Error(t, err)

	// Valid Author Update
	updated, err := svc.UpdateArticle(ctx, command.UpdateArticleCommand{
		ID:      articleDTO.ID,
		UserID:  authorID,
		Title:   "Clean Architecture in Go - 2nd Edition",
		Content: "Updated content.",
		Status:  entity.ArticleStatusPublished,
	})
	require.NoError(t, err)
	assert.Equal(t, "clean-architecture-in-go-2nd-edition", updated.Slug)
	assert.Equal(t, string(entity.ArticleStatusPublished), updated.Status)

	// 5. Delete Article - Ownership Guard
	err = svc.DeleteArticle(ctx, command.DeleteArticleCommand{
		ID:     articleDTO.ID,
		UserID: strangerID, // Stranger cannot delete
	})
	assert.Error(t, err)

	// Valid Author Delete
	err = svc.DeleteArticle(ctx, command.DeleteArticleCommand{
		ID:     articleDTO.ID,
		UserID: authorID,
	})
	require.NoError(t, err)

	_, err = svc.GetArticle(ctx, query.GetArticleByIDQuery{ID: articleDTO.ID})
	assert.Error(t, err)
}
