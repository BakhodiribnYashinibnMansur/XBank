package query

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListResult struct {
	Items      []*repository.SystemErrorView `json:"items"`
	Pagination domain.Pagination             `json:"pagination"`
}

type ListHandler struct{ pool *pgxpool.Pool }

func NewListHandler(pool *pgxpool.Pool) *ListHandler { return &ListHandler{pool: pool} }

func (h *ListHandler) Handle(ctx context.Context, filter repository.SystemErrorFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	countQ := `SELECT COUNT(*) FROM system_errors WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if filter.Severity != "" {
		countQ += fmt.Sprintf(` AND severity = $%d`, idx)
		args = append(args, filter.Severity)
		idx++
	}
	if filter.Resolution != "" {
		countQ += fmt.Sprintf(` AND resolution = $%d`, idx)
		args = append(args, filter.Resolution)
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

	listQ := `SELECT id, code, message, severity, category, request_id, ip_address, path, method, resolution,
	                 COALESCE(resolved_by,''), to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	          FROM system_errors WHERE 1=1`
	lArgs := []interface{}{}
	lIdx := 1

	if filter.Severity != "" {
		listQ += fmt.Sprintf(` AND severity = $%d`, lIdx)
		lArgs = append(lArgs, filter.Severity)
		lIdx++
	}
	if filter.Resolution != "" {
		listQ += fmt.Sprintf(` AND resolution = $%d`, lIdx)
		lArgs = append(lArgs, filter.Resolution)
		lIdx++
	}
	if filter.Code != "" {
		listQ += fmt.Sprintf(` AND code ILIKE $%d`, lIdx)
		lArgs = append(lArgs, "%"+filter.Code+"%")
		lIdx++
	}

	listQ += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, lIdx, lIdx+1)
	lArgs = append(lArgs, filter.Limit, filter.Offset)

	rows, err := h.pool.Query(ctx, listQ, lArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*repository.SystemErrorView
	for rows.Next() {
		v := &repository.SystemErrorView{}
		if err := rows.Scan(&v.ID, &v.Code, &v.Message, &v.Severity, &v.Category,
			&v.RequestID, &v.IPAddress, &v.Path, &v.Method, &v.Resolution, &v.ResolvedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}

	return &ListResult{Items: items, Pagination: domain.NewPagination(total, filter.Limit, filter.Offset)}, nil
}
