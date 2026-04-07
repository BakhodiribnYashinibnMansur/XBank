package domain

import (
	"context"

)

// WriteRepository defines write operations for Translation aggregate.
type WriteRepository interface {
	Save(ctx context.Context, t *Translation) error
	Update(ctx context.Context, t *Translation) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Translation, error)
	FindByKeyAndLanguage(ctx context.Context, key string, lang Language) (*Translation, error)
}

// TranslationView is the read projection.
type TranslationView struct {
	ID       string          `json:"id"`
	Key      string          `json:"key"`
	Language Language `json:"language"`
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
