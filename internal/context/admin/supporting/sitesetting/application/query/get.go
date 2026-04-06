package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/repository"
)

// GetHandler fetches a single site setting by ID.
type GetHandler struct {
	readRepo repository.ReadRepository
}

func NewGetHandler(readRepo repository.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*repository.SiteSettingView, error) {
	return h.readRepo.FindByID(ctx, id)
}
