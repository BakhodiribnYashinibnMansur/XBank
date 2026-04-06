package query

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListResult struct {
	Items      []*repository.ErrorCodeView `json:"items"`
	Pagination domain.Pagination           `json:"pagination"`
}

type ListHandler struct{ pool *pgxpool.Pool }

func NewListHandler(pool *pgxpool.Pool) *ListHandler { return &ListHandler{pool: pool} }

func (h *ListHandler) Handle(ctx context.Context, filter repository.ErrorCodeFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	countQ := `SELECT COUNT(*) FROM error_codes WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if filter.Category != "" {
		countQ += fmt.Sprintf(` AND category = $%d`, idx)
		args = append(args, filter.Category)
		idx++
	}
	if filter.Severity != "" {
		countQ += fmt.Sprintf(` AND severity = $%d`, idx)
		args = append(args, filter.Severity)
		idx++
	}
	if filter.Code != "" {
		countQ += fmt.Sprintf(` AND code ILIKE $%d`, idx)
		args = append(args, "%"+filter.Code+"%")
		idx++
	}

	var total int64
	if err := h.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	listQ := `SELECT id, code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion
	          FROM error_codes WHERE 1=1`
	lArgs := []interface{}{}
	lIdx := 1
	if filter.Category != "" {
		listQ += fmt.Sprintf(` AND category = $%d`, lIdx)
		lArgs = append(lArgs, filter.Category)
		lIdx++
	}
	if filter.Severity != "" {
		listQ += fmt.Sprintf(` AND severity = $%d`, lIdx)
		lArgs = append(lArgs, filter.Severity)
		lIdx++
	}
	if filter.Code != "" {
		listQ += fmt.Sprintf(` AND code ILIKE $%d`, lIdx)
		lArgs = append(lArgs, "%"+filter.Code+"%")
		lIdx++
	}
	listQ += fmt.Sprintf(` ORDER BY code LIMIT $%d OFFSET $%d`, lIdx, lIdx+1)
	lArgs = append(lArgs, filter.Limit, filter.Offset)

	rows, err := h.pool.Query(ctx, listQ, lArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*repository.ErrorCodeView
	for rows.Next() {
		v := &repository.ErrorCodeView{}
		if err := rows.Scan(&v.ID, &v.Code, &v.MessageEn, &v.MessageUz, &v.MessageRu,
			&v.Category, &v.Severity, &v.HTTPStatus, &v.Retryable, &v.Suggestion); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return &ListResult{Items: items, Pagination: domain.NewPagination(total, filter.Limit, filter.Offset)}, nil
}
