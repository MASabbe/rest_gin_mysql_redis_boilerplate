package response

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/dto"
)

type ArticleResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromArticleDTO(dto *dto.ArticleDTO) ArticleResponse {
	if dto == nil {
		return ArticleResponse{}
	}
	return ArticleResponse{
		ID:        dto.ID,
		UserID:    dto.UserID,
		Title:     dto.Title,
		Slug:      dto.Slug,
		Content:   dto.Content,
		Status:    dto.Status,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func FromArticleDTOList(dtos []dto.ArticleDTO) []ArticleResponse {
	list := make([]ArticleResponse, 0, len(dtos))
	for _, d := range dtos {
		list = append(list, FromArticleDTO(&d))
	}
	return list
}
