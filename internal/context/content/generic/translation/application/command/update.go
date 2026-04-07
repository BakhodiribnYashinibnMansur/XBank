package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

type UpdateHandler struct {
	repo     domain.WriteRepository
	eventBus appKernel.EventBus
}

func NewUpdateHandler(repo domain.WriteRepository, bus appKernel.EventBus) *UpdateHandler {
	return &UpdateHandler{repo: repo, eventBus: bus}
}

func (h *UpdateHandler) Handle(ctx context.Context, id, value string) error {
	t, err := h.repo.FindByID(ctx, id)
	if err != nil || t == nil {
		return domain.ErrTranslationNotFound
	}

	if err := t.Update(value); err != nil {
		return fmt.Errorf("update translation: %w", err)
	}

	if err := h.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("update translation: save: %w", err)
	}

	h.eventBus.Publish(ctx, domain.NewTranslationUpdated(t.ID, t.Key, string(t.Language)))
	return nil
}
