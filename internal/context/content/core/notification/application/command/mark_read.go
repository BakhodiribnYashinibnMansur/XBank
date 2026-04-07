package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain"
)

type MarkReadHandler struct {
	repo domain.WriteRepository
}

func NewMarkReadHandler(repo domain.WriteRepository) *MarkReadHandler {
	return &MarkReadHandler{repo: repo}
}

func (h *MarkReadHandler) Handle(ctx context.Context, id string) error {
	n, err := h.repo.FindByID(ctx, id)
	if err != nil || n == nil {
		return domain.ErrNotificationNotFound
	}
	n.MarkAsRead()
	return h.repo.Update(ctx, n)
}
