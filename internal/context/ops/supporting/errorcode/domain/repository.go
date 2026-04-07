package domain

import (
	"context"

)

type WriteRepository interface {
	Save(ctx context.Context, e *ErrorCode) error
	Update(ctx context.Context, e *ErrorCode) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*ErrorCode, error)
	FindByCode(ctx context.Context, code string) (*ErrorCode, error)
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
