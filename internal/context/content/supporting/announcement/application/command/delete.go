package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
)

type DeleteHandler struct{ repo domain.WriteRepository }

func NewDeleteHandler(repo domain.WriteRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	if _, err := h.repo.FindByID(ctx, id); err != nil {
		return domain.ErrAnnouncementNotFound
	}
	return h.repo.Delete(ctx, id)
}
