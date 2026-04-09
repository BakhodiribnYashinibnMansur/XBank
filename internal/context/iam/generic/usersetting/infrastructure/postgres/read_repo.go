package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) List(ctx context.Context, filter domain.UserSettingFilter) ([]*domain.UserSettingView, int64, error) {
	start := time.Now()

	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if filter.UserID != "" {
		where += fmt.Sprintf(" AND user_id=$%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Key != "" {
		where += fmt.Sprintf(" AND key=$%d", argIdx)
		args = append(args, filter.Key)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM user_settings "+where, args...).Scan(&total); err != nil {
		metrics.ObserveQuery("UserSettingReadRepo.List.Count", start, err)
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, key, value FROM user_settings %s ORDER BY key LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		metrics.ObserveQuery("UserSettingReadRepo.List", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.UserSettingView
	for rows.Next() {
		v := &domain.UserSettingView{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.Key, &v.Value); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("UserSettingReadRepo.List", start, rows.Err())
	return items, total, rows.Err()
}
