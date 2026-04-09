package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.FileView, error) {
	start := time.Now()
	v := &domain.FileView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, original_name, mime_type, size, url, COALESCE(uploaded_by,''), created_at::text
		 FROM files WHERE id=$1`, id,
	).Scan(&v.ID, &v.Name, &v.OriginalName, &v.MimeType, &v.Size, &v.URL, &v.UploadedBy, &v.CreatedAt)
	metrics.ObserveQuery("FileReadRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.FileFilter) ([]*domain.FileView, int64, error) {
	start := time.Now()

	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if filter.MimeType != "" {
		where += fmt.Sprintf(" AND mime_type=$%d", argIdx)
		args = append(args, filter.MimeType)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM files "+where, args...).Scan(&total); err != nil {
		metrics.ObserveQuery("FileReadRepo.List.Count", start, err)
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(
		`SELECT id, name, original_name, mime_type, size, url, COALESCE(uploaded_by,''), created_at::text
		 FROM files %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		metrics.ObserveQuery("FileReadRepo.List", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.FileView
	for rows.Next() {
		v := &domain.FileView{}
		if err := rows.Scan(&v.ID, &v.Name, &v.OriginalName, &v.MimeType, &v.Size, &v.URL, &v.UploadedBy, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("FileReadRepo.List", start, rows.Err())
	return items, total, rows.Err()
}
