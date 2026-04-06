package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ListResult struct {
	Items      []*repository.TranslationView `json:"items"`
	Pagination domain.Pagination             `json:"pagination"`
}

type ListHandler struct {
	readRepo repository.ReadRepository
}

func NewListHandler(repo repository.ReadRepository) *ListHandler {
	return &ListHandler{readRepo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, filter repository.TranslationFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	items, total, err := h.readRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Items:      items,
		Pagination: domain.NewPagination(total, filter.Limit, filter.Offset),
	}, nil
}
