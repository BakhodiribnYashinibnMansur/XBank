package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

type CreateHandler struct {
	repo     repository.WriteRepository
	eventBus appKernel.EventBus
}

func NewCreateHandler(repo repository.WriteRepository, bus appKernel.EventBus) *CreateHandler {
	return &CreateHandler{repo: repo, eventBus: bus}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateNotificationRequest) (string, error) {
	n, err := entity.NewNotification(req.UserID, req.Title, req.Message, entity.NotificationType(req.Type), req.Data)
	if err != nil {
		return "", fmt.Errorf("create notification: %w", err)
	}
	if err := h.repo.Save(ctx, n); err != nil {
		return "", fmt.Errorf("create notification: save: %w", err)
	}
	h.eventBus.Publish(ctx, event.NewNotificationSent(n.ID, n.UserID, n.Title, string(n.Type)))
	return n.ID, nil
}
