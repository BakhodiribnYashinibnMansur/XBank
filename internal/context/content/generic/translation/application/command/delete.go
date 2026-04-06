package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
)

type DeleteHandler struct {
	repo repository.WriteRepository
}

func NewDeleteHandler(repo repository.WriteRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	if _, err := h.repo.FindByID(ctx, id); err != nil {
		return entity.ErrTranslationNotFound
	}
	return h.repo.Delete(ctx, id)
}
