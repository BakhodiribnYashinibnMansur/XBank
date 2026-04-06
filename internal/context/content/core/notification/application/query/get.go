package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/repository"
)

type GetHandler struct{ repo repository.ReadRepository }

func NewGetHandler(r repository.ReadRepository) *GetHandler { return &GetHandler{repo: r} }

func (h *GetHandler) Handle(ctx context.Context, id string) (*repository.NotificationView, error) {
	return h.repo.FindByID(ctx, id)
}
