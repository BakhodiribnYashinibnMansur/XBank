package repository

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/entity"
)

// WriteRepository defines write operations for Translation aggregate.
type WriteRepository interface {
	Save(ctx context.Context, t *entity.Translation) error
	Update(ctx context.Context, t *entity.Translation) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Translation, error)
	FindByKeyAndLanguage(ctx context.Context, key string, lang entity.Language) (*entity.Translation, error)
}

// TranslationView is the read projection.
type TranslationView struct {
	ID       string          `json:"id"`
	Key      string          `json:"key"`
	Language entity.Language `json:"language"`
	Value    string          `json:"value"`
	Group    string          `json:"group"`
}

// TranslationFilter for list queries.
type TranslationFilter struct {
	Language string
	Group    string
	Key      string
	Limit    int
	Offset   int
}

// ReadRepository defines read operations.
type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*TranslationView, error)
	List(ctx context.Context, filter TranslationFilter) ([]*TranslationView, int64, error)
}
