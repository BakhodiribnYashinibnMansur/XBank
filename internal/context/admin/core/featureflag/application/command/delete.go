package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

// DeleteHandler deletes a feature flag.
type DeleteHandler struct {
	writeRepo domain.WriteRepository
	eventBus  appKernel.EventBus
}

func NewDeleteHandler(repo domain.WriteRepository, bus appKernel.EventBus) *DeleteHandler {
	return &DeleteHandler{writeRepo: repo, eventBus: bus}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	flag, err := h.writeRepo.FindByID(ctx, id)
	if err != nil || flag == nil {
		return domain.ErrFlagNotFound
	}

	if err := h.writeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}

	h.eventBus.Publish(ctx, domain.NewFlagDeleted(flag.ID, flag.Key))
	return nil
}
