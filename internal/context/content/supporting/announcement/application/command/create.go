package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
)

type CreateHandler struct{ repo domain.WriteRepository }

func NewCreateHandler(repo domain.WriteRepository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateAnnouncementRequest) (string, error) {
	a, err := domain.NewAnnouncement(req.TitleUz, req.TitleRu, req.TitleEn, req.BodyUz, req.BodyRu, req.BodyEn, req.Priority)
	if err != nil {
		return "", fmt.Errorf("create announcement: %w", err)
	}
	a.StartDate = req.StartDate
	a.EndDate = req.EndDate

	if err := h.repo.Save(ctx, a); err != nil {
		return "", fmt.Errorf("create announcement: save: %w", err)
	}
	return a.ID, nil
}
