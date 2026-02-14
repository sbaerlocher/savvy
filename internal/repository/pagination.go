// Package repository defines data access interfaces.
package repository

import (
	"math"

	"gorm.io/gorm"
)

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page    int // 1-based page number
	PerPage int // Items per page
}

// PaginatedResult holds paginated query results.
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// DefaultPerPage is the default number of items per page.
const DefaultPerPage = 25

// MaxPerPage is the maximum allowed items per page.
// This prevents excessive memory usage and keeps response sizes reasonable.
const MaxPerPage = 100

// NormalizePagination ensures page and per_page have valid values.
func NormalizePagination(params PaginationParams) PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = DefaultPerPage
	}
	if params.PerPage > MaxPerPage {
		params.PerPage = MaxPerPage
	}
	return params
}

// ApplyPagination adds LIMIT and OFFSET to a GORM query.
func ApplyPagination(query *gorm.DB, params PaginationParams) *gorm.DB {
	offset := (params.Page - 1) * params.PerPage
	return query.Offset(offset).Limit(params.PerPage)
}

// CalculateTotalPages calculates the total number of pages.
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}
