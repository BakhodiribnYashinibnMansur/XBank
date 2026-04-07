package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

type CreateHandler struct {
	repo     domain.WriteRepository
	eventBus appKernel.EventBus
}

func NewCreateHandler(repo domain.WriteRepository, bus appKernel.EventBus) *CreateHandler {
	return &CreateHandler{repo: repo, eventBus: bus}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateNotificationRequest) (string, error) {
	n, err := domain.NewNotification(req.UserID, req.Title, req.Message, domain.NotificationType(req.Type), req.Data)
	if err != nil {
		return "", fmt.Errorf("create notification: %w", err)
	}
	if err := h.repo.Save(ctx, n); err != nil {
		return "", fmt.Errorf("create notification: save: %w", err)
	}
	h.eventBus.Publish(ctx, domain.NewNotificationSent(n.ID, n.UserID, n.Title, string(n.Type)))
	return n.ID, nil
}
