package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/repository"
)

type MarkReadHandler struct {
	repo repository.WriteRepository
}

func NewMarkReadHandler(repo repository.WriteRepository) *MarkReadHandler {
	return &MarkReadHandler{repo: repo}
}

func (h *MarkReadHandler) Handle(ctx context.Context, id string) error {
	n, err := h.repo.FindByID(ctx, id)
	if err != nil || n == nil {
		return entity.ErrNotificationNotFound
	}
	n.MarkAsRead()
	return h.repo.Update(ctx, n)
}
