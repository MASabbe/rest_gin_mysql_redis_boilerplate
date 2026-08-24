package request

type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=255"`
	Content string `json:"content" binding:"required"`
}

type UpdateArticleRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3,max=255"`
	Content string `json:"content"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published archived"`
}
