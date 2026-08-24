package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
	"github.com/go-sql-driver/mysql"
)

type MySQLArticleRepository struct {
	db *sql.DB
}

// NewMySQLArticleRepository creates a new MySQL Article repository.
func NewMySQLArticleRepository(db *sql.DB) repository.ArticleRepository {
	return &MySQLArticleRepository{db: db}
}

func (r *MySQLArticleRepository) Create(ctx context.Context, article *entity.Article) error {
	query := `
		INSERT INTO articles (id, user_id, title, slug, content, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		article.ID,
		article.UserID,
		article.Title,
		article.Slug,
		article.Content,
		string(article.Status),
		article.CreatedAt,
		article.UpdatedAt,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("article with slug '%s' already exists", article.Slug), err)
		}
		return appErrors.NewInternalError("failed to create article", err)
	}

	return nil
}

func (r *MySQLArticleRepository) FindByID(ctx context.Context, id string) (*entity.Article, error) {
	query := `
		SELECT id, user_id, title, slug, content, status, created_at, updated_at
		FROM articles
		WHERE id = ?
		LIMIT 1
	`
	var a entity.Article
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID,
		&a.UserID,
		&a.Title,
		&a.Slug,
		&a.Content,
		&statusStr,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("article not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find article by id", err)
	}

	a.Status = entity.ArticleStatus(statusStr)
	return &a, nil
}

func (r *MySQLArticleRepository) FindBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	query := `
		SELECT id, user_id, title, slug, content, status, created_at, updated_at
		FROM articles
		WHERE slug = ?
		LIMIT 1
	`
	var a entity.Article
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(slug))).Scan(
		&a.ID,
		&a.UserID,
		&a.Title,
		&a.Slug,
		&a.Content,
		&statusStr,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("article not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find article by slug", err)
	}

	a.Status = entity.ArticleStatus(statusStr)
	return &a, nil
}

func (r *MySQLArticleRepository) List(ctx context.Context, filter repository.ArticleFilter, p pagination.Pagination, s queryparam.Sorting) ([]*entity.Article, int64, error) {
	whereClauses := []string{"1=1"}
	var args []any

	if filter.UserID != "" {
		whereClauses = append(whereClauses, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, string(filter.Status))
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "title LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Total count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM articles WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, appErrors.NewInternalError("failed to count articles", err)
	}

	// Safe order by
	orderBy := s.SQLOrderBy()
	if strings.TrimSpace(s.Field) == "" {
		orderBy = "created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, title, slug, content, status, created_at, updated_at
		FROM articles
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereSQL, orderBy)

	queryArgs := append(args, p.Limit(), p.Offset())
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list articles", err)
	}
	defer rows.Close()

	var articles []*entity.Article
	for rows.Next() {
		var a entity.Article
		var statusStr string
		if err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Title,
			&a.Slug,
			&a.Content,
			&statusStr,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, 0, appErrors.NewInternalError("failed to scan article row", err)
		}
		a.Status = entity.ArticleStatus(statusStr)
		articles = append(articles, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, appErrors.NewInternalError("error iterating articles", err)
	}

	return articles, total, nil
}

func (r *MySQLArticleRepository) Update(ctx context.Context, article *entity.Article) error {
	query := `
		UPDATE articles
		SET title = ?, slug = ?, content = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Slug,
		article.Content,
		string(article.Status),
		article.UpdatedAt,
		article.ID,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("article with slug '%s' already exists", article.Slug), err)
		}
		return appErrors.NewInternalError("failed to update article", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("article not found for update")
	}

	return nil
}

func (r *MySQLArticleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.NewInternalError("failed to delete article", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("article not found for deletion")
	}

	return nil
}

var _ repository.ArticleRepository = (*MySQLArticleRepository)(nil)
