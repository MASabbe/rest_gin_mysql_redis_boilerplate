package command

import (
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type CreateArticleCommand struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (c CreateArticleCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return appErrors.NewValidationError("author user ID is required", appErrors.FieldError{
			Field:   "user_id",
			Message: "cannot be empty",
		})
	}
	if strings.TrimSpace(c.Title) == "" {
		return appErrors.NewValidationError("article title is required", appErrors.FieldError{
			Field:   "title",
			Message: "cannot be empty",
		})
	}
	if strings.TrimSpace(c.Content) == "" {
		return appErrors.NewValidationError("article content is required", appErrors.FieldError{
			Field:   "content",
			Message: "cannot be empty",
		})
	}
	return nil
}

type UpdateArticleCommand struct {
	ID      string               `json:"id"`
	UserID  string               `json:"user_id"` // For author ownership check
	Title   string               `json:"title"`
	Content string               `json:"content"`
	Status  entity.ArticleStatus `json:"status"`
}

func (c UpdateArticleCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("article ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type DeleteArticleCommand struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"` // For author ownership check
}

func (c DeleteArticleCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("article ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}
