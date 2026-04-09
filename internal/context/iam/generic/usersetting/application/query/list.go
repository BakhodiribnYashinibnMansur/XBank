package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
)

type ListResult struct {
	Items []*domain.UserSettingView `json:"items"`
	Total int64                     `json:"total"`
}

type ListHandler struct {
	repo domain.ReadRepository
}

func NewListHandler(repo domain.ReadRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, filter domain.UserSettingFilter) (*ListResult, error) {
	items, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ListResult{Items: items, Total: total}, nil
}
