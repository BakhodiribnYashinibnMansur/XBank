package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/google/uuid"
)

// RequestHandler creates a new data export request.
type RequestHandler struct {
	repo domain.WriteRepository
}

func NewRequestHandler(repo domain.WriteRepository) *RequestHandler {
	return &RequestHandler{repo: repo}
}

func (h *RequestHandler) Handle(ctx context.Context, req application.RequestExportRequest) (string, error) {
	var err error
	defer metrics.ObserveService("DataExportService", "RequestExport", time.Now(), &err)

	export, err := domain.NewDataExport(req.UserID)
	if err != nil {
		return "", err
	}
	export.ID = uuid.New().String()

	if err = h.repo.Save(ctx, export); err != nil {
		return "", err
	}
	return export.ID, nil
}
