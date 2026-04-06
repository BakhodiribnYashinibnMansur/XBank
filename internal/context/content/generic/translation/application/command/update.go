package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

type UpdateHandler struct {
	repo     repository.WriteRepository
	eventBus appKernel.EventBus
}

func NewUpdateHandler(repo repository.WriteRepository, bus appKernel.EventBus) *UpdateHandler {
	return &UpdateHandler{repo: repo, eventBus: bus}
}

func (h *UpdateHandler) Handle(ctx context.Context, id, value string) error {
	t, err := h.repo.FindByID(ctx, id)
	if err != nil || t == nil {
		return entity.ErrTranslationNotFound
	}

	if err := t.Update(value); err != nil {
		return fmt.Errorf("update translation: %w", err)
	}

	if err := h.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("update translation: save: %w", err)
	}

	h.eventBus.Publish(ctx, event.NewTranslationUpdated(t.ID, t.Key, string(t.Language)))
	return nil
}
