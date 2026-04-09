package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
)

// ListResult contains paginated export results.
type ListResult struct {
	Items []*domain.DataExportView `json:"items"`
	Total int64                    `json:"total"`
}

// ListHandler retrieves a paginated list of data exports.
type ListHandler struct {
	repo domain.ReadRepository
}

func NewListHandler(repo domain.ReadRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, filter domain.DataExportFilter) (*ListResult, error) {
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}
