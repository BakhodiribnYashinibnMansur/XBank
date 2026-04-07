package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain"
)

type UpdateHandler struct{ repo domain.WriteRepository }

func NewUpdateHandler(r domain.WriteRepository) *UpdateHandler { return &UpdateHandler{repo: r} }

func (h *UpdateHandler) Handle(ctx context.Context, id string, req application.UpdateErrorCodeRequest) error {
	e, err := h.repo.FindByID(ctx, id)
	if err != nil || e == nil {
		return domain.ErrCodeNotFound
	}
	e.Update(req.MessageEn, req.MessageUz, req.MessageRu, req.Suggestion, req.HTTPStatus, req.Retryable)
	if err := h.repo.Update(ctx, e); err != nil {
		return fmt.Errorf("update error code: %w", err)
	}
	return nil
}
