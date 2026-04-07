package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

// CreateHandler creates a new feature flag.
type CreateHandler struct {
	writeRepo domain.WriteRepository
	eventBus  appKernel.EventBus
}

func NewCreateHandler(repo domain.WriteRepository, bus appKernel.EventBus) *CreateHandler {
	return &CreateHandler{writeRepo: repo, eventBus: bus}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateFlagRequest) (string, error) {
	// Check key uniqueness
	if existing, _ := h.writeRepo.FindByKey(ctx, req.Key); existing != nil {
		return "", domain.ErrKeyExists
	}

	flag, err := domain.NewFeatureFlag(req.Key, req.Description, req.FlagType, req.DefaultValue)
	if err != nil {
		return "", fmt.Errorf("create flag: %w", err)
	}

	if err := h.writeRepo.Save(ctx, flag); err != nil {
		return "", fmt.Errorf("create flag: save: %w", err)
	}

	h.eventBus.Publish(ctx, domain.NewFlagCreated(flag.ID, flag.Key))
	return flag.ID, nil
}
