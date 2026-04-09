package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.SystemErrorView, error) {
	start := time.Now()
	v := &domain.SystemErrorView{}
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, code, message, severity, category, COALESCE(stack_trace,''),
		        COALESCE(request_id,''), COALESCE(user_id,''), COALESCE(ip_address,''),
		        COALESCE(path,''), COALESCE(method,''), COALESCE(metadata,'{}'),
		        resolution, COALESCE(resolved_by,''), created_at::text
		 FROM system_errors WHERE id=$1`, id,
	).Scan(&v.ID, &v.Code, &v.Message, &v.Severity, &v.Category,
		&v.StackTrace, &v.RequestID, &v.UserID, &v.IPAddress,
		&v.Path, &v.Method, &metadataJSON,
		&v.Resolution, &v.ResolvedBy, &v.CreatedAt)
	metrics.ObserveQuery("SystemErrorReadRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &v.Metadata)
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.SystemErrorFilter) ([]*domain.SystemErrorView, int64, error) {
	start := time.Now()

	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if filter.Code != "" {
		where += fmt.Sprintf(" AND code=$%d", argIdx)
		args = append(args, filter.Code)
		argIdx++
	}
	if filter.Severity != "" {
		where += fmt.Sprintf(" AND severity=$%d", argIdx)
		args = append(args, filter.Severity)
		argIdx++
	}
	if filter.Resolution != "" {
		where += fmt.Sprintf(" AND resolution=$%d", argIdx)
		args = append(args, filter.Resolution)
		argIdx++
	}
	if filter.DateFrom != "" {
		where += fmt.Sprintf(" AND created_at >= $%d::timestamptz", argIdx)
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		where += fmt.Sprintf(" AND created_at <= $%d::timestamptz", argIdx)
		args = append(args, filter.DateTo)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM system_errors "+where, args...).Scan(&total); err != nil {
		metrics.ObserveQuery("SystemErrorReadRepo.List.Count", start, err)
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(
		`SELECT id, code, message, severity, category, COALESCE(stack_trace,''),
		        COALESCE(request_id,''), COALESCE(user_id,''), COALESCE(ip_address,''),
		        COALESCE(path,''), COALESCE(method,''), COALESCE(metadata,'{}'),
		        resolution, COALESCE(resolved_by,''), created_at::text
		 FROM system_errors %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		metrics.ObserveQuery("SystemErrorReadRepo.List", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.SystemErrorView
	for rows.Next() {
		v := &domain.SystemErrorView{}
		var metadataJSON []byte
		if err := rows.Scan(&v.ID, &v.Code, &v.Message, &v.Severity, &v.Category,
			&v.StackTrace, &v.RequestID, &v.UserID, &v.IPAddress,
			&v.Path, &v.Method, &metadataJSON,
			&v.Resolution, &v.ResolvedBy, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &v.Metadata)
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("SystemErrorReadRepo.List", start, rows.Err())
	return items, total, rows.Err()
}
