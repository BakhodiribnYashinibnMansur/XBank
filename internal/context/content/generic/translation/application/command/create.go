package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
)

type CreateHandler struct {
	repo repository.WriteRepository
}

func NewCreateHandler(repo repository.WriteRepository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateTranslationRequest) (string, error) {
	lang := entity.Language(req.Language)

	if existing, _ := h.repo.FindByKeyAndLanguage(ctx, req.Key, lang); existing != nil {
		return "", entity.ErrKeyLanguageExists
	}

	t, err := entity.NewTranslation(req.Key, lang, req.Value, req.Group)
	if err != nil {
		return "", fmt.Errorf("create translation: %w", err)
	}

	if err := h.repo.Save(ctx, t); err != nil {
		return "", fmt.Errorf("create translation: save: %w", err)
	}
	return t.ID, nil
}
