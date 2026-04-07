package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
	kernelDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// ListResult holds paginated feature flags.
type ListResult struct {
	Items      []*domain.FeatureFlagView `json:"items"`
	Pagination kernelDomain.Pagination             `json:"pagination"`
}

// ListHandler returns paginated feature flags.
type ListHandler struct {
	readRepo domain.ReadRepository
}

func NewListHandler(readRepo domain.ReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

func (h *ListHandler) Handle(ctx context.Context, filter domain.FeatureFlagFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
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
