package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

// UpdateHandler updates an existing feature flag.
type UpdateHandler struct {
	writeRepo repository.WriteRepository
	eventBus  appKernel.EventBus
}

func NewUpdateHandler(repo repository.WriteRepository, bus appKernel.EventBus) *UpdateHandler {
	return &UpdateHandler{writeRepo: repo, eventBus: bus}
}

func (h *UpdateHandler) Handle(ctx context.Context, id string, req application.UpdateFlagRequest) error {
	flag, err := h.writeRepo.FindByID(ctx, id)
	if err != nil || flag == nil {
		return entity.ErrFlagNotFound
	}

	if err := flag.Update(req.Description, req.DefaultValue, req.Active, req.RolloutPct); err != nil {
		return fmt.Errorf("update flag: %w", err)
	}

	if err := h.writeRepo.Update(ctx, flag); err != nil {
		return fmt.Errorf("update flag: save: %w", err)
	}

	h.eventBus.Publish(ctx, event.NewFlagUpdated(flag.ID, flag.Key))
	return nil
}
