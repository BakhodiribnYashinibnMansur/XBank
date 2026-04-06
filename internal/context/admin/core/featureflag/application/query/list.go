package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// ListResult holds paginated feature flags.
type ListResult struct {
	Items      []*repository.FeatureFlagView `json:"items"`
	Pagination domain.Pagination             `json:"pagination"`
}

// ListHandler returns paginated feature flags.
type ListHandler struct {
	readRepo repository.ReadRepository
}

func NewListHandler(readRepo repository.ReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

func (h *ListHandler) Handle(ctx context.Context, filter repository.FeatureFlagFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
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
