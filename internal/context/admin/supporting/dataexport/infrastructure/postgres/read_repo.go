package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.DataExportView, error) {
	start := time.Now()
	v := &domain.DataExportView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, COALESCE(file_url,''), COALESCE(error_msg,''), created_at::text
		 FROM data_exports WHERE id=$1`, id,
	).Scan(&v.ID, &v.UserID, &v.Status, &v.FileURL, &v.ErrorMsg, &v.CreatedAt)
	metrics.ObserveQuery("DataExportReadRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.DataExportFilter) ([]*domain.DataExportView, int64, error) {
	start := time.Now()

	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if filter.UserID != "" {
		where += fmt.Sprintf(" AND user_id=$%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM data_exports " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		metrics.ObserveQuery("DataExportReadRepo.List.Count", start, err)
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset

	query := fmt.Sprintf(
		`SELECT id, user_id, status, COALESCE(file_url,''), COALESCE(error_msg,''), created_at::text
		 FROM data_exports %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		metrics.ObserveQuery("DataExportReadRepo.List", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.DataExportView
	for rows.Next() {
		v := &domain.DataExportView{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.Status, &v.FileURL, &v.ErrorMsg, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("DataExportReadRepo.List", start, rows.Err())
	return items, total, rows.Err()
}
