package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

type PublishHandler struct {
	repo     repository.WriteRepository
	eventBus appKernel.EventBus
}

func NewPublishHandler(repo repository.WriteRepository, bus appKernel.EventBus) *PublishHandler {
	return &PublishHandler{repo: repo, eventBus: bus}
}

func (h *PublishHandler) Handle(ctx context.Context, id string) error {
	a, err := h.repo.FindByID(ctx, id)
	if err != nil || a == nil {
		return entity.ErrAnnouncementNotFound
	}
	if err := a.Publish(); err != nil {
		return fmt.Errorf("publish announcement: %w", err)
	}
	if err := h.repo.Update(ctx, a); err != nil {
		return fmt.Errorf("publish announcement: save: %w", err)
	}
	h.eventBus.Publish(ctx, event.NewAnnouncementPublished(a.ID))
	return nil
}
