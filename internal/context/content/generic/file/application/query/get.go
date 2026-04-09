package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
)

type GetHandler struct {
	repo domain.ReadRepository
}

func NewGetHandler(repo domain.ReadRepository) *GetHandler {
	return &GetHandler{repo: repo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*domain.FileView, error) {
	view, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return nil, domain.ErrFileNotFound
	}
	return view, nil
}
