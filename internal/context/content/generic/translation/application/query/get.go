package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
)

type GetHandler struct {
	readRepo repository.ReadRepository
}

func NewGetHandler(repo repository.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: repo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*repository.TranslationView, error) {
	return h.readRepo.FindByID(ctx, id)
}
