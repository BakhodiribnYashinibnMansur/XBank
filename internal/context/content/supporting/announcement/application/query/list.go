package query

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ListResult struct {
	Items      []*repository.AnnouncementView `json:"items"`
	Pagination domain.Pagination              `json:"pagination"`
}

type ListHandler struct{ repo repository.ReadRepository }

func NewListHandler(r repository.ReadRepository) *ListHandler { return &ListHandler{repo: r} }

func (h *ListHandler) Handle(ctx context.Context, filter repository.AnnouncementFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Pagination: domain.NewPagination(total, filter.Limit, filter.Offset)}, nil
}

type ListActiveHandler struct{ repo repository.ReadRepository }

func NewListActiveHandler(r repository.ReadRepository) *ListActiveHandler {
	return &ListActiveHandler{repo: r}
}

func (h *ListActiveHandler) Handle(ctx context.Context) ([]*repository.AnnouncementView, error) {
	return h.repo.ListActive(ctx, time.Now())
}
