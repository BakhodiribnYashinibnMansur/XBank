package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/event"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/repository"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
)

// UpsertHandler creates or updates a site setting by key.
type UpsertHandler struct {
	writeRepo repository.WriteRepository
	eventBus  appKernel.EventBus
}

func NewUpsertHandler(writeRepo repository.WriteRepository, eventBus appKernel.EventBus) *UpsertHandler {
	return &UpsertHandler{writeRepo: writeRepo, eventBus: eventBus}
}

func (h *UpsertHandler) Handle(ctx context.Context, req application.CreateSettingRequest) (*application.SettingResponse, error) {
	existing, _ := h.writeRepo.FindByKey(ctx, req.Key)
	if existing != nil {
		// Update existing
		existing.Update(&req.Value, &req.Description)
		if err := h.writeRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("upsert setting: update: %w", err)
		}
		h.eventBus.Publish(ctx, event.NewSettingUpdated(existing.ID, existing.Key, existing.Value))
		return toResponse(existing), nil
	}

	// Create new
	setting, err := entity.NewSiteSetting(req.Key, req.Value, req.SettingType, req.Description)
	if err != nil {
		return nil, fmt.Errorf("upsert setting: create: %w", err)
	}

	if err := h.writeRepo.Save(ctx, setting); err != nil {
		return nil, fmt.Errorf("upsert setting: save: %w", err)
	}

	h.eventBus.Publish(ctx, event.NewSettingCreated(setting.ID, setting.Key))
	return toResponse(setting), nil
}

func toResponse(s *entity.SiteSetting) *application.SettingResponse {
	return &application.SettingResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       s.Value,
		SettingType: string(s.SettingType),
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
