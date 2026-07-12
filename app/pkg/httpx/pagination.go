package httpx

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Pagination is the response-side pagination metadata.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// PageParams are the parsed, validated request-side pagination inputs.
type PageParams struct {
	Page     int
	PageSize int
}

// Offset returns the SQL OFFSET for the requested page.
func (p PageParams) Offset() int32 { return int32((p.Page - 1) * p.PageSize) }

// Limit returns the SQL LIMIT for the requested page.
func (p PageParams) Limit() int32 { return int32(p.PageSize) }

// BuildPagination assembles response metadata from a total row count.
func (p PageParams) BuildPagination(total int64) Pagination {
	totalPages := int64(0)
	if p.PageSize > 0 {
		totalPages = (total + int64(p.PageSize) - 1) / int64(p.PageSize)
	}
	return Pagination{Page: p.Page, PageSize: p.PageSize, Total: total, TotalPages: totalPages}
}

// ParsePageParams reads ?page & ?page_size, clamping to safe bounds.
func ParsePageParams(c *gin.Context) PageParams {
	page := atoiDefault(c.Query("page"), defaultPage)
	if page < 1 {
		page = defaultPage
	}
	size := atoiDefault(c.Query("page_size"), defaultPageSize)
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return PageParams{Page: page, PageSize: size}
}

// Sort represents a validated sort instruction.
type Sort struct {
	Field string
	Desc  bool
}

// ParseSort reads ?sort=field:dir and validates the field against an allow-list,
// returning the fallback when the input is empty or not permitted.
func ParseSort(c *gin.Context, allowed map[string]string, fallback Sort) Sort {
	raw := strings.TrimSpace(c.Query("sort"))
	if raw == "" {
		return fallback
	}
	field, dir, _ := strings.Cut(raw, ":")
	col, ok := allowed[strings.ToLower(strings.TrimSpace(field))]
	if !ok {
		return fallback
	}
	return Sort{Field: col, Desc: strings.EqualFold(strings.TrimSpace(dir), "desc")}
}

// Order renders the sort as a SQL ORDER BY direction keyword.
func (s Sort) Order() string {
	if s.Desc {
		return "DESC"
	}
	return "ASC"
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
