package query

import (
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
)

type GetArticleByIDQuery struct {
	ID string `json:"id"`
}

func (q GetArticleByIDQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return appErrors.NewValidationError("article ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type GetArticleBySlugQuery struct {
	Slug string `json:"slug"`
}

func (q GetArticleBySlugQuery) Validate() error {
	if strings.TrimSpace(q.Slug) == "" {
		return appErrors.NewValidationError("article slug is required", appErrors.FieldError{
			Field:   "slug",
			Message: "cannot be empty",
		})
	}
	return nil
}

type ListArticlesQuery struct {
	UserID     string
	Status     entity.ArticleStatus
	Search     string
	Pagination pagination.Pagination
	Sorting    queryparam.Sorting
}
