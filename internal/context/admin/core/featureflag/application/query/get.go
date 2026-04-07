package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
)

// GetHandler fetches a feature flag by ID with rule groups.
type GetHandler struct {
	readRepo domain.ReadRepository
}

func NewGetHandler(readRepo domain.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*domain.FeatureFlagView, error) {
	return h.readRepo.FindByID(ctx, id)
}
