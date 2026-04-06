package domain

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// NewPagination creates pagination metadata from query results.
func NewPagination(total int64, limit, offset int) Pagination {
	return Pagination{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+limit) < total,
	}
}
