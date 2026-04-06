package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/repository"
)

type CreateHandler struct{ repo repository.WriteRepository }

func NewCreateHandler(r repository.WriteRepository) *CreateHandler { return &CreateHandler{repo: r} }

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateErrorCodeRequest) (string, error) {
	if existing, _ := h.repo.FindByCode(ctx, req.Code); existing != nil {
		return "", entity.ErrCodeAlreadyExists
	}
	e, err := entity.NewErrorCode(req.Code, req.MessageEn, req.MessageUz, req.MessageRu,
		req.Category, req.Severity, req.HTTPStatus, req.Retryable, req.Suggestion)
	if err != nil {
		return "", fmt.Errorf("create error code: %w", err)
	}
	if err := h.repo.Save(ctx, e); err != nil {
		return "", fmt.Errorf("create error code: save: %w", err)
	}
	return e.ID, nil
}
