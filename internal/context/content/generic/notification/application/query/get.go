package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
)

type GetHandler struct{ repo domain.ReadRepository }

func NewGetHandler(r domain.ReadRepository) *GetHandler { return &GetHandler{repo: r} }

func (h *GetHandler) Handle(ctx context.Context, id string) (*domain.NotificationView, error) {
	return h.repo.FindByID(ctx, id)
}
