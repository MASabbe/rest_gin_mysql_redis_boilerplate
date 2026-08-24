package queryparam

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// Sorting represents safe sorting parameters.
type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"` // "ASC" or "DESC"
}

// SQLOrderBy returns a safe SQL clause, e.g. "created_at DESC".
func (s Sorting) SQLOrderBy() string {
	return fmt.Sprintf("%s %s", s.Field, s.Order)
}

// ExtractSorting extracts and validates sort parameters against an explicit allowlist.
// If the requested field is not in the allowlist, defaultField and defaultOrder are applied.
func ExtractSorting(c *gin.Context, allowlist []string, defaultField, defaultOrder string) Sorting {
	field := strings.ToLower(strings.TrimSpace(c.Query("sort")))
	order := strings.ToUpper(strings.TrimSpace(c.Query("order")))

	// Normalize order
	if order != "ASC" && order != "DESC" {
		if strings.ToUpper(defaultOrder) == "ASC" {
			order = "ASC"
		} else {
			order = "DESC"
		}
	}

	// Validate field against allowlist
	isAllowed := false
	for _, allowed := range allowlist {
		if strings.EqualFold(field, allowed) {
			field = strings.ToLower(allowed)
			isAllowed = true
			break
		}
	}

	if !isAllowed || field == "" {
		field = defaultField
	}

	return Sorting{
		Field: field,
		Order: order,
	}
}
