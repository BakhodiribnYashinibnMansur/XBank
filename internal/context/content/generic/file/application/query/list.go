package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
)

type ListResult struct {
	Items []*domain.FileView `json:"items"`
	Total int64              `json:"total"`
}

type ListHandler struct {
	repo domain.ReadRepository
}

func NewListHandler(repo domain.ReadRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, filter domain.FileFilter) (*ListResult, error) {
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}
