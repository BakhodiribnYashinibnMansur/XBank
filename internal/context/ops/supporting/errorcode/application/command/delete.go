package command

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain"
)

type DeleteHandler struct{ repo domain.WriteRepository }

func NewDeleteHandler(r domain.WriteRepository) *DeleteHandler { return &DeleteHandler{repo: r} }

func (h *DeleteHandler) Handle(ctx context.Context, id string) error {
	if _, err := h.repo.FindByID(ctx, id); err != nil {
		return domain.ErrCodeNotFound
	}
	return h.repo.Delete(ctx, id)
}
