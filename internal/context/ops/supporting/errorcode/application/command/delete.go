package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/repository"
)

type DeleteHandler struct{ repo repository.WriteRepository }

func NewDeleteHandler(r repository.WriteRepository) *DeleteHandler { return &DeleteHandler{repo: r} }

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	if _, err := h.repo.FindByID(ctx, id); err != nil {
		return entity.ErrCodeNotFound
	}
	return h.repo.Delete(ctx, id)
}
