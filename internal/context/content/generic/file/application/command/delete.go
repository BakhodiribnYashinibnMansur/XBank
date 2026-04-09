package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// DeleteHandler handles file deletion.
type DeleteHandler struct {
	repo domain.WriteRepository
}

func NewDeleteHandler(repo domain.WriteRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("FileService", "Delete", time.Now(), &err)
	return h.repo.Delete(ctx, id)
}
