package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/infrastructure/persistence"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLArticleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLArticleRepository(db)
	ctx := context.Background()

	article, _ := entity.NewArticle("user-1", "Article Title", "Article content here")

	// 1. Success
	mock.ExpectExec("INSERT INTO articles").
		WithArgs(article.ID, article.UserID, article.Title, article.Slug, article.Content, string(article.Status), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, article)
	assert.NoError(t, err)

	// 2. Duplicate Slug Conflict
	mock.ExpectExec("INSERT INTO articles").
		WithArgs(article.ID, article.UserID, article.Title, article.Slug, article.Content, string(article.Status), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("Error 1062 (23000): Duplicate entry 'article-title' for key 'slug'"))

	err = repo.Create(ctx, article)
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeConflict, appErr.Type)
}

func TestMySQLArticleRepository_FindByID_And_BySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLArticleRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	// 1. FindByID - Found
	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "slug", "content", "status", "created_at", "updated_at"}).
		AddRow("art-123", "user-1", "Title", "slug-1", "Content", "published", now, now)

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs("art-123").
		WillReturnRows(rows)

	art, err := repo.FindByID(ctx, "art-123")
	assert.NoError(t, err)
	assert.Equal(t, "art-123", art.ID)

	// 2. FindByID - Not Found
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs("art-unknown").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindByID(ctx, "art-unknown")
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeNotFound, appErr.Type)

	// 3. FindBySlug - Found
	rowsSlug := sqlmock.NewRows([]string{"id", "user_id", "title", "slug", "content", "status", "created_at", "updated_at"}).
		AddRow("art-123", "user-1", "Title", "slug-1", "Content", "published", now, now)

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE slug = ?").
		WithArgs("slug-1").
		WillReturnRows(rowsSlug)

	artSlug, err := repo.FindBySlug(ctx, "slug-1")
	assert.NoError(t, err)
	assert.Equal(t, "slug-1", artSlug.Slug)
}

func TestMySQLArticleRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLArticleRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM articles WHERE (.+)").
		WithArgs("published", "%golang%").
		WillReturnRows(countRows)

	// Mock select query
	selectRows := sqlmock.NewRows([]string{"id", "user_id", "title", "slug", "content", "status", "created_at", "updated_at"}).
		AddRow("art-1", "user-1", "Golang Architecture", "golang-architecture", "Content", "published", now, now)

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE (.+) ORDER BY (.+) LIMIT \\? OFFSET \\?").
		WithArgs("published", "%golang%", 10, 0).
		WillReturnRows(selectRows)

	filter := repository.ArticleFilter{
		Status: entity.ArticleStatusPublished,
		Search: "golang",
	}
	articles, total, err := repo.List(ctx, filter, pagination.Pagination{Page: 1, PageSize: 10}, queryparam.Sorting{Field: "created_at", Order: "DESC"})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, articles, 1)
	assert.Equal(t, "golang-architecture", articles[0].Slug)
}

func TestMySQLArticleRepository_Update_And_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLArticleRepository(db)
	ctx := context.Background()

	article, _ := entity.NewArticle("user-1", "Updated Title", "Updated Content")

	// 1. Update - Success
	mock.ExpectExec("UPDATE articles SET").
		WithArgs(article.Title, article.Slug, article.Content, string(article.Status), sqlmock.AnyArg(), article.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(ctx, article)
	assert.NoError(t, err)

	// 2. Update - Not Found
	mock.ExpectExec("UPDATE articles SET").
		WithArgs(article.Title, article.Slug, article.Content, string(article.Status), sqlmock.AnyArg(), article.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(ctx, article)
	assert.Error(t, err)
	assert.Equal(t, appErrors.TypeNotFound, appErrors.AsAppError(err).Type)

	// 3. Delete - Success
	mock.ExpectExec("DELETE FROM articles WHERE id = ?").
		WithArgs(article.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, article.ID)
	assert.NoError(t, err)

	// 4. Delete - Not Found
	mock.ExpectExec("DELETE FROM articles WHERE id = ?").
		WithArgs("unknown-art-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(ctx, "unknown-art-id")
	assert.Error(t, err)
	assert.Equal(t, appErrors.TypeNotFound, appErrors.AsAppError(err).Type)
}
