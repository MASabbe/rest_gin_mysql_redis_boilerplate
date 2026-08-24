package application

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/dto"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

// ArticleService defines the application use case contracts for articles.
type ArticleService interface {
	CreateArticle(ctx context.Context, cmd command.CreateArticleCommand) (*dto.ArticleDTO, error)
	GetArticle(ctx context.Context, q query.GetArticleByIDQuery) (*dto.ArticleDTO, error)
	GetArticleBySlug(ctx context.Context, q query.GetArticleBySlugQuery) (*dto.ArticleDTO, error)
	ListArticles(ctx context.Context, q query.ListArticlesQuery) ([]dto.ArticleDTO, pagination.Meta, error)
	UpdateArticle(ctx context.Context, cmd command.UpdateArticleCommand) (*dto.ArticleDTO, error)
	DeleteArticle(ctx context.Context, cmd command.DeleteArticleCommand) error
}

type articleService struct {
	articleRepo repository.ArticleRepository
}

// NewArticleService creates a new ArticleService instance.
func NewArticleService(articleRepo repository.ArticleRepository) ArticleService {
	return &articleService{articleRepo: articleRepo}
}

func (s *articleService) CreateArticle(ctx context.Context, cmd command.CreateArticleCommand) (*dto.ArticleDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	article, err := entity.NewArticle(cmd.UserID, cmd.Title, cmd.Content)
	if err != nil {
		return nil, err
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, err
	}

	res := dto.ToArticleDTO(article)
	return &res, nil
}

func (s *articleService) GetArticle(ctx context.Context, q query.GetArticleByIDQuery) (*dto.ArticleDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	article, err := s.articleRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	res := dto.ToArticleDTO(article)
	return &res, nil
}

func (s *articleService) GetArticleBySlug(ctx context.Context, q query.GetArticleBySlugQuery) (*dto.ArticleDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	article, err := s.articleRepo.FindBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	res := dto.ToArticleDTO(article)
	return &res, nil
}

func (s *articleService) ListArticles(ctx context.Context, q query.ListArticlesQuery) ([]dto.ArticleDTO, pagination.Meta, error) {
	filter := repository.ArticleFilter{
		UserID: q.UserID,
		Status: q.Status,
		Search: q.Search,
	}

	articles, total, err := s.articleRepo.List(ctx, filter, q.Pagination, q.Sorting)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	meta := pagination.NewMeta(q.Pagination.Page, q.Pagination.PageSize, total)
	return dto.ToArticleDTOList(articles), meta, nil
}

func (s *articleService) UpdateArticle(ctx context.Context, cmd command.UpdateArticleCommand) (*dto.ArticleDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	article, err := s.articleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// Author ownership check (if specified in command)
	if cmd.UserID != "" && article.UserID != cmd.UserID {
		return nil, appErrors.NewForbiddenError("you can only modify your own articles")
	}

	if err := article.Update(cmd.Title, cmd.Content, cmd.Status); err != nil {
		return nil, err
	}

	if err := s.articleRepo.Update(ctx, article); err != nil {
		return nil, err
	}

	res := dto.ToArticleDTO(article)
	return &res, nil
}

func (s *articleService) DeleteArticle(ctx context.Context, cmd command.DeleteArticleCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	article, err := s.articleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// Author ownership check (if specified in command)
	if cmd.UserID != "" && article.UserID != cmd.UserID {
		return appErrors.NewForbiddenError("you can only delete your own articles")
	}

	return s.articleRepo.Delete(ctx, cmd.ID)
}
