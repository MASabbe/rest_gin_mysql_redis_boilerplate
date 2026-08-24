package repository

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
)

// ArticleFilter contains query filters for listing articles.
type ArticleFilter struct {
	UserID string               `json:"user_id"`
	Status entity.ArticleStatus `json:"status"`
	Search string               `json:"search"`
}

// ArticleRepository defines persistence contracts for Article entities.
type ArticleRepository interface {
	Create(ctx context.Context, article *entity.Article) error
	FindByID(ctx context.Context, id string) (*entity.Article, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Article, error)
	List(ctx context.Context, filter ArticleFilter, p pagination.Pagination, s queryparam.Sorting) ([]*entity.Article, int64, error)
	Update(ctx context.Context, article *entity.Article) error
	Delete(ctx context.Context, id string) error
}
