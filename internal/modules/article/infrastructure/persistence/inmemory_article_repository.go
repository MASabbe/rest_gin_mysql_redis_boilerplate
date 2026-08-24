package persistence

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
)

type InMemoryArticleRepository struct {
	mu             sync.RWMutex
	articles       map[string]*entity.Article
	articlesBySlug map[string]*entity.Article
}

func NewInMemoryArticleRepository() *InMemoryArticleRepository {
	return &InMemoryArticleRepository{
		articles:       make(map[string]*entity.Article),
		articlesBySlug: make(map[string]*entity.Article),
	}
}

func (r *InMemoryArticleRepository) Create(ctx context.Context, article *entity.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cleanSlug := strings.ToLower(article.Slug)
	if _, exists := r.articlesBySlug[cleanSlug]; exists {
		return appErrors.NewConflictError(fmt.Sprintf("article with slug '%s' already exists", article.Slug))
	}

	r.articles[article.ID] = article
	r.articlesBySlug[cleanSlug] = article
	return nil
}

func (r *InMemoryArticleRepository) FindByID(ctx context.Context, id string) (*entity.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	article, exists := r.articles[id]
	if !exists {
		return nil, appErrors.NewNotFoundError("article not found")
	}
	return article, nil
}

func (r *InMemoryArticleRepository) FindBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	article, exists := r.articlesBySlug[strings.ToLower(slug)]
	if !exists {
		return nil, appErrors.NewNotFoundError("article not found")
	}
	return article, nil
}

func (r *InMemoryArticleRepository) List(ctx context.Context, filter repository.ArticleFilter, p pagination.Pagination, s queryparam.Sorting) ([]*entity.Article, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entity.Article
	for _, a := range r.articles {
		if filter.UserID != "" && a.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && a.Status != filter.Status {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(a.Title), strings.ToLower(filter.Search)) {
			continue
		}
		filtered = append(filtered, a)
	}

	total := int64(len(filtered))
	offset := p.Offset()
	limit := p.Limit()

	if offset >= len(filtered) {
		return []*entity.Article{}, total, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

func (r *InMemoryArticleRepository) Update(ctx context.Context, article *entity.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.articles[article.ID]
	if !exists {
		return appErrors.NewNotFoundError("article not found")
	}

	cleanSlug := strings.ToLower(article.Slug)
	if other, slugTaken := r.articlesBySlug[cleanSlug]; slugTaken && other.ID != article.ID {
		return appErrors.NewConflictError(fmt.Sprintf("article with slug '%s' already exists", article.Slug))
	}

	delete(r.articlesBySlug, strings.ToLower(existing.Slug))
	r.articles[article.ID] = article
	r.articlesBySlug[cleanSlug] = article
	return nil
}

func (r *InMemoryArticleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	article, exists := r.articles[id]
	if !exists {
		return appErrors.NewNotFoundError("article not found")
	}

	delete(r.articles, id)
	delete(r.articlesBySlug, strings.ToLower(article.Slug))
	return nil
}

var _ repository.ArticleRepository = (*InMemoryArticleRepository)(nil)
