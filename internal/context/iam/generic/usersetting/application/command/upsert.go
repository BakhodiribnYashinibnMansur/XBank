package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/google/uuid"
)

// UpsertHandler creates or updates a user setting.
type UpsertHandler struct {
	repo domain.WriteRepository
}

func NewUpsertHandler(repo domain.WriteRepository) *UpsertHandler {
	return &UpsertHandler{repo: repo}
}

func (h *UpsertHandler) Handle(ctx context.Context, req application.UpsertSettingRequest) (err error) {
	defer metrics.ObserveService("UserSettingService", "Upsert", time.Now(), &err)

	setting, err := domain.NewUserSetting(req.UserID, req.Key, req.Value)
	if err != nil {
		return err
	}
	setting.ID = uuid.New().String()

	return h.repo.Upsert(ctx, setting)
}
