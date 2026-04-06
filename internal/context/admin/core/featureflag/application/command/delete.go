package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

// DeleteHandler deletes a feature flag.
type DeleteHandler struct {
	writeRepo repository.WriteRepository
	eventBus  appKernel.EventBus
}

func NewDeleteHandler(repo repository.WriteRepository, bus appKernel.EventBus) *DeleteHandler {
	return &DeleteHandler{writeRepo: repo, eventBus: bus}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	flag, err := h.writeRepo.FindByID(ctx, id)
	if err != nil || flag == nil {
		return entity.ErrFlagNotFound
	}

	if err := h.writeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}

	h.eventBus.Publish(ctx, event.NewFlagDeleted(flag.ID, flag.Key))
	return nil
}
