package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain"
)

// GetHandler fetches a single site setting by ID.
type GetHandler struct {
	readRepo domain.ReadRepository
}

func NewGetHandler(readRepo domain.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*domain.SiteSettingView, error) {
	return h.readRepo.FindByID(ctx, id)
}
