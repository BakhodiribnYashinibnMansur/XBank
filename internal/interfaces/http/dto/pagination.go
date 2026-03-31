package dto

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PaginationRequest - query params: ?page=1&limit=20
type PaginationRequest struct {
	Page  int
	Limit int
}

// Offset - SQL OFFSET
func (p PaginationRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginationResponse - response metadata
type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// PaginatedResponse - generic paginated response envelope
type PaginatedResponse struct {
	Data       any                `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// ParsePagination - extracts page/limit from query params with validation
func ParsePagination(c *fiber.Ctx) PaginationRequest {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return PaginationRequest{Page: page, Limit: limit}
}
