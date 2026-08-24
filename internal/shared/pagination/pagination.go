package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Pagination represents query parameters for paginated requests.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Offset returns the SQL OFFSET for query execution.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the SQL LIMIT for query execution.
func (p Pagination) Limit() int {
	return p.PageSize
}

// Meta holds pagination metadata for API responses.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta calculates pagination metadata.
func NewMeta(page, pageSize int, totalCount int64) Meta {
	totalPages := 0
	if pageSize > 0 && totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	return Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

// Extract extracts and normalizes pagination parameters from Gin query parameters.
func Extract(c *gin.Context) Pagination {
	page := DefaultPage
	pageSize := DefaultPageSize

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > MaxPageSize {
				pageSize = MaxPageSize
			} else {
				pageSize = ps
			}
		}
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}
