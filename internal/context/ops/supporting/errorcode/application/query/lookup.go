package query

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LookupHandler struct{ pool *pgxpool.Pool }

func NewLookupHandler(pool *pgxpool.Pool) *LookupHandler { return &LookupHandler{pool: pool} }

func (h *LookupHandler) Handle(ctx context.Context, code string) (*repository.ErrorCodeView, error) {
	v := &repository.ErrorCodeView{}
	err := h.pool.QueryRow(ctx,
		`SELECT id, code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion
		 FROM error_codes WHERE code = $1`, code,
	).Scan(&v.ID, &v.Code, &v.MessageEn, &v.MessageUz, &v.MessageRu,
		&v.Category, &v.Severity, &v.HTTPStatus, &v.Retryable, &v.Suggestion)
	if err != nil {
		return nil, fmt.Errorf("error code lookup: %w", err)
	}
	return v, nil
}
