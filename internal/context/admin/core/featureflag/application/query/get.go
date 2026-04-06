package query

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
)

// GetHandler fetches a feature flag by ID with rule groups.
type GetHandler struct {
	readRepo repository.ReadRepository
}

func NewGetHandler(readRepo repository.ReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

func (h *GetHandler) Handle(ctx context.Context, id string) (*repository.FeatureFlagView, error) {
	return h.readRepo.FindByID(ctx, id)
}
