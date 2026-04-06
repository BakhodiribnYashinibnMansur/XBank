package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/repository"
)

type DeleteHandler struct{ repo repository.WriteRepository }

func NewDeleteHandler(repo repository.WriteRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	if _, err := h.repo.FindByID(ctx, id); err != nil {
		return entity.ErrAnnouncementNotFound
	}
	return h.repo.Delete(ctx, id)
}
