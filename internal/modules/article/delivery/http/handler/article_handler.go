package handler

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/delivery/http/request"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/delivery/http/response"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/domain/entity"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/queryparam"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

var allowedArticleSortFields = []string{"created_at", "title", "updated_at", "status"}

type ArticleHandler struct {
	articleService application.ArticleService
}

// NewArticleHandler creates a new ArticleHandler instance.
func NewArticleHandler(articleService application.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleService: articleService}
}

// Create handles POST /api/v1/articles
func (h *ArticleHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		sharedResponse.Error(c, appErrors.NewUnauthorizedError("unauthorized"))
		return
	}

	var req request.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.articleService.CreateArticle(c.Request.Context(), command.CreateArticleCommand{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Created(c, "Article created successfully", response.FromArticleDTO(result))
}

// GetByID handles GET /api/v1/articles/:id
func (h *ArticleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.articleService.GetArticle(c.Request.Context(), query.GetArticleByIDQuery{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Article retrieved successfully", response.FromArticleDTO(result))
}

// GetBySlug handles GET /api/v1/articles/slug/:slug
func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	result, err := h.articleService.GetArticleBySlug(c.Request.Context(), query.GetArticleBySlugQuery{Slug: slug})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Article retrieved successfully", response.FromArticleDTO(result))
}

// List handles GET /api/v1/articles
func (h *ArticleHandler) List(c *gin.Context) {
	p := pagination.Extract(c)
	s := queryparam.ExtractSorting(c, allowedArticleSortFields, "created_at", "DESC")

	userIDFilter := c.Query("user_id")
	statusFilter := entity.ArticleStatus(c.Query("status"))
	searchFilter := c.Query("search")

	list, meta, err := h.articleService.ListArticles(c.Request.Context(), query.ListArticlesQuery{
		UserID:     userIDFilter,
		Status:     statusFilter,
		Search:     searchFilter,
		Pagination: p,
		Sorting:    s,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Success(c, 200, "Articles retrieved successfully", response.FromArticleDTOList(list), meta)
}

// Update handles PUT /api/v1/articles/:id
func (h *ArticleHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetAuthenticatedUserID(c)
	id := c.Param("id")

	var req request.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.articleService.UpdateArticle(c.Request.Context(), command.UpdateArticleCommand{
		ID:      id,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Status:  entity.ArticleStatus(req.Status),
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Article updated successfully", response.FromArticleDTO(result))
}

// Delete handles DELETE /api/v1/articles/:id
func (h *ArticleHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetAuthenticatedUserID(c)
	id := c.Param("id")

	err := h.articleService.DeleteArticle(c.Request.Context(), command.DeleteArticleCommand{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Article deleted successfully", nil)
}
