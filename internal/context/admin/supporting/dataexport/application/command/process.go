package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// ProcessHandler starts processing an export and completes or fails it.
type ProcessHandler struct {
	repo domain.WriteRepository
}

func NewProcessHandler(repo domain.WriteRepository) *ProcessHandler {
	return &ProcessHandler{repo: repo}
}

func (h *ProcessHandler) StartProcessing(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("DataExportService", "StartProcessing", time.Now(), &err)

	export, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if export == nil {
		return domain.ErrExportNotFound
	}
	if err = export.StartProcessing(); err != nil {
		return err
	}
	return h.repo.Update(ctx, export)
}

func (h *ProcessHandler) Complete(ctx context.Context, id, fileURL string) (err error) {
	defer metrics.ObserveService("DataExportService", "Complete", time.Now(), &err)

	export, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if export == nil {
		return domain.ErrExportNotFound
	}
	if err = export.Complete(fileURL); err != nil {
		return err
	}
	return h.repo.Update(ctx, export)
}

func (h *ProcessHandler) Fail(ctx context.Context, id, reason string) (err error) {
	defer metrics.ObserveService("DataExportService", "Fail", time.Now(), &err)

	export, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if export == nil {
		return domain.ErrExportNotFound
	}
	if err = export.Fail(reason); err != nil {
		return err
	}
	return h.repo.Update(ctx, export)
}
