package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
)

type GetHandler struct {
	readRepo domain.ReadRepository
}

func NewGetHandler(repo domain.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: repo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*domain.TranslationView, error) {
	return h.readRepo.FindByID(ctx, id)
}
