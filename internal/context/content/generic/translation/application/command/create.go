package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
)

type CreateHandler struct {
	repo domain.WriteRepository
}

func NewCreateHandler(repo domain.WriteRepository) *CreateHandler {
	return &CreateHandler{repo: repo}
}

func (h *CreateHandler) Handle(ctx context.Context, req application.CreateTranslationRequest) (string, error) {
	lang := domain.Language(req.Language)

	if existing, _ := h.repo.FindByKeyAndLanguage(ctx, req.Key, lang); existing != nil {
		return "", domain.ErrKeyLanguageExists
	}

	t, err := domain.NewTranslation(req.Key, lang, req.Value, req.Group)
	if err != nil {
		return "", fmt.Errorf("create translation: %w", err)
	}

	if err := h.repo.Save(ctx, t); err != nil {
		return "", fmt.Errorf("create translation: save: %w", err)
	}
	return t.ID, nil
}
