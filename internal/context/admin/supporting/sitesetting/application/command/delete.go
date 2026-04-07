package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain"
)

// DeleteHandler deletes a site setting by ID.
type DeleteHandler struct {
	writeRepo domain.WriteRepository
}

func NewDeleteHandler(writeRepo domain.WriteRepository) *DeleteHandler {
	return &DeleteHandler{writeRepo: writeRepo}
}

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	existing, err := h.writeRepo.FindByID(ctx, id)
	if err != nil || existing == nil {
		return domain.ErrSettingNotFound
	}
	if err := h.writeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete setting: %w", err)
	}
	return nil
}
