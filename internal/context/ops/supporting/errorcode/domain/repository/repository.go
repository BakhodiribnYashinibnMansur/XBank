package repository

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/entity"
)

type WriteRepository interface {
	Save(ctx context.Context, e *entity.ErrorCode) error
	Update(ctx context.Context, e *entity.ErrorCode) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.ErrorCode, error)
	FindByCode(ctx context.Context, code string) (*entity.ErrorCode, error)
}

type ErrorCodeView struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	MessageEn  string `json:"message_en"`
	MessageUz  string `json:"message_uz"`
	MessageRu  string `json:"message_ru"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	HTTPStatus int    `json:"http_status"`
	Retryable  bool   `json:"retryable"`
	Suggestion string `json:"suggestion"`
}

type ErrorCodeFilter struct {
	Code     string
	Category string
	Severity string
	Limit    int
	Offset   int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*ErrorCodeView, error)
	FindByCode(ctx context.Context, code string) (*ErrorCodeView, error)
	List(ctx context.Context, filter ErrorCodeFilter) ([]*ErrorCodeView, int64, error)
}
