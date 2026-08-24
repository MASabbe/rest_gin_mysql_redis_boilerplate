package dto

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
)

type ArticleDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToArticleDTO(a *entity.Article) ArticleDTO {
	if a == nil {
		return ArticleDTO{}
	}
	return ArticleDTO{
		ID:        a.ID,
		UserID:    a.UserID,
		Title:     a.Title,
		Slug:      a.Slug,
		Content:   a.Content,
		Status:    string(a.Status),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func ToArticleDTOList(articles []*entity.Article) []ArticleDTO {
	list := make([]ArticleDTO, 0, len(articles))
	for _, a := range articles {
		list = append(list, ToArticleDTO(a))
	}
	return list
}
