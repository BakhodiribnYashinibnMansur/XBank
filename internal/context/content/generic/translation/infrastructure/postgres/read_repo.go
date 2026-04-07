package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.TranslationView, error) {
	v := &domain.TranslationView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, language, value, "group" FROM translations WHERE id = $1`, id,
	).Scan(&v.ID, &v.Key, &v.Language, &v.Value, &v.Group)
	if err != nil {
		return nil, fmt.Errorf("translation read: %w", err)
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.TranslationFilter) ([]*domain.TranslationView, int64, error) {
	countQ := `SELECT COUNT(*) FROM translations WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if filter.Language != "" {
		countQ += fmt.Sprintf(` AND language = $%d`, idx)
		args = append(args, filter.Language)
		idx++
	}
	if filter.Group != "" {
		countQ += fmt.Sprintf(` AND "group" = $%d`, idx)
		args = append(args, filter.Group)
		idx++
	}
	if filter.Key != "" {
		countQ += fmt.Sprintf(` AND key ILIKE $%d`, idx)
		args = append(args, "%"+filter.Key+"%")
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `SELECT id, key, language, value, "group" FROM translations WHERE 1=1`
	lArgs := []interface{}{}
	lIdx := 1

	if filter.Language != "" {
		listQ += fmt.Sprintf(` AND language = $%d`, lIdx)
		lArgs = append(lArgs, filter.Language)
		lIdx++
	}
	if filter.Group != "" {
		listQ += fmt.Sprintf(` AND "group" = $%d`, lIdx)
		lArgs = append(lArgs, filter.Group)
		lIdx++
	}
	if filter.Key != "" {
		listQ += fmt.Sprintf(` AND key ILIKE $%d`, lIdx)
		lArgs = append(lArgs, "%"+filter.Key+"%")
		lIdx++
	}

	listQ += fmt.Sprintf(` ORDER BY key, language LIMIT $%d OFFSET $%d`, lIdx, lIdx+1)
	lArgs = append(lArgs, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQ, lArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.TranslationView
	for rows.Next() {
		v := &domain.TranslationView{}
		if err := rows.Scan(&v.ID, &v.Key, &v.Language, &v.Value, &v.Group); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}
