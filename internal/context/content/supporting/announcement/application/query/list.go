package query

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
	kernelDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ListResult struct {
	Items      []*domain.AnnouncementView `json:"items"`
	Pagination kernelDomain.Pagination              `json:"pagination"`
}

type ListHandler struct{ repo domain.ReadRepository }

func NewListHandler(r domain.ReadRepository) *ListHandler { return &ListHandler{repo: r} }

func (h *ListHandler) Handle(ctx context.Context, filter domain.AnnouncementFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Pagination: kernelDomain.NewPagination(total, filter.Limit, filter.Offset)}, nil
}

type ListActiveHandler struct{ repo domain.ReadRepository }

func NewListActiveHandler(r domain.ReadRepository) *ListActiveHandler {
	return &ListActiveHandler{repo: r}
}

func (h *ListActiveHandler) Handle(ctx context.Context) ([]*domain.AnnouncementView, error) {
	return h.repo.ListActive(ctx, time.Now())
}
