package command

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/google/uuid"
)

// UploadHandler handles file metadata creation.
type UploadHandler struct {
	repo domain.WriteRepository
}

func NewUploadHandler(repo domain.WriteRepository) *UploadHandler {
	return &UploadHandler{repo: repo}
}

func (h *UploadHandler) Handle(ctx context.Context, originalName, mimeType string, size int64, path, url, uploadedBy string) (_ string, err error) {
	defer metrics.ObserveService("FileService", "Upload", time.Now(), &err)

	name := uuid.New().String()
	file, err := domain.NewFile(name, originalName, mimeType, size, path, url, uploadedBy)
	if err != nil {
		return "", err
	}
	file.ID = uuid.New().String()

	if err = h.repo.Save(ctx, file); err != nil {
		return "", fmt.Errorf("save file metadata: %w", err)
	}
	return file.ID, nil
}
