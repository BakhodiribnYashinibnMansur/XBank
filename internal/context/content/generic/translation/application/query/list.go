package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
	kernelDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ListResult struct {
	Items      []*domain.TranslationView `json:"items"`
	Pagination kernelDomain.Pagination             `json:"pagination"`
}

type ListHandler struct {
	readRepo domain.ReadRepository
}

func NewListHandler(repo domain.ReadRepository) *ListHandler {
	return &ListHandler{readRepo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, filter domain.TranslationFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	items, total, err := h.readRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Items:      items,
		Pagination: kernelDomain.NewPagination(total, filter.Limit, filter.Offset),
	}, nil
}
