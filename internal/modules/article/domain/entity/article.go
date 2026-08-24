package entity

import (
	"regexp"
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/google/uuid"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
)

// Article represents the Article domain entity.
type Article struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"` // Author
	Title     string        `json:"title"`
	Slug      string        `json:"slug"`
	Content   string        `json:"content"`
	Status    ArticleStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewArticle creates and validates a new Article entity in draft status.
func NewArticle(userID, title, content string) (*Article, error) {
	cleanUserID := strings.TrimSpace(userID)
	if cleanUserID == "" {
		return nil, appErrors.NewValidationError("author user ID is required", appErrors.FieldError{
			Field:   "user_id",
			Message: "cannot be empty",
		})
	}

	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return nil, appErrors.NewValidationError("article title is required", appErrors.FieldError{
			Field:   "title",
			Message: "cannot be empty",
		})
	}

	if len(cleanTitle) < 3 || len(cleanTitle) > 255 {
		return nil, appErrors.NewValidationError("article title must be between 3 and 255 characters", appErrors.FieldError{
			Field:   "title",
			Message: "length must be 3-255 characters",
		})
	}

	cleanContent := strings.TrimSpace(content)
	if cleanContent == "" {
		return nil, appErrors.NewValidationError("article content is required", appErrors.FieldError{
			Field:   "content",
			Message: "cannot be empty",
		})
	}

	now := time.Now().UTC()
	slug := GenerateSlug(cleanTitle)

	return &Article{
		ID:        uuid.New().String(),
		UserID:    cleanUserID,
		Title:     cleanTitle,
		Slug:      slug,
		Content:   cleanContent,
		Status:    ArticleStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update updates the mutable fields of the article.
func (a *Article) Update(title, content string, status ArticleStatus) error {
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle != "" {
		if len(cleanTitle) < 3 || len(cleanTitle) > 255 {
			return appErrors.NewValidationError("article title must be between 3 and 255 characters", appErrors.FieldError{
				Field:   "title",
				Message: "length must be 3-255 characters",
			})
		}
		a.Title = cleanTitle
		a.Slug = GenerateSlug(cleanTitle)
	}

	cleanContent := strings.TrimSpace(content)
	if cleanContent != "" {
		a.Content = cleanContent
	}

	if status != "" {
		if !isValidStatus(status) {
			return appErrors.NewValidationError("invalid article status", appErrors.FieldError{
				Field:   "status",
				Message: "must be draft, published, or archived",
			})
		}
		a.Status = status
	}

	a.UpdatedAt = time.Now().UTC()
	return nil
}

func isValidStatus(status ArticleStatus) bool {
	switch status {
	case ArticleStatusDraft, ArticleStatusPublished, ArticleStatusArchived:
		return true
	default:
		return false
	}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateSlug generates a clean URL slug from a title.
func GenerateSlug(title string) string {
	lowered := strings.ToLower(strings.TrimSpace(title))
	slug := nonAlphanumericRegex.ReplaceAllString(lowered, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = uuid.New().String()[:8]
	}
	return slug
}
