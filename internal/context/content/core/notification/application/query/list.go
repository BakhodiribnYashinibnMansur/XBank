package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain"
	kernelDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ListResult struct {
	Items      []*domain.NotificationView `json:"items"`
	Pagination kernelDomain.Pagination              `json:"pagination"`
}

type ListHandler struct{ repo domain.ReadRepository }

func NewListHandler(r domain.ReadRepository) *ListHandler { return &ListHandler{repo: r} }

func (h *ListHandler) Handle(ctx context.Context, filter domain.NotificationFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Items:      items,
		Pagination: kernelDomain.NewPagination(total, filter.Limit, filter.Offset),
	}, nil
}
