package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReadRepo implements domain.ReadRepository for PostgreSQL.
type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.SiteSettingView, error) {
	v := &domain.SiteSettingView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, value, setting_type, description,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at,
		        to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at
		 FROM site_settings WHERE id = $1`, id,
	).Scan(&v.ID, &v.Key, &v.Value, &v.SettingType, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("site_setting read find by id: %w", err)
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.SiteSettingFilter) ([]*domain.SiteSettingView, int64, error) {
	// Count
	countQuery := `SELECT COUNT(*) FROM site_settings WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Key != "" {
		countQuery += fmt.Sprintf(` AND key ILIKE $%d`, argIdx)
		args = append(args, "%"+filter.Key+"%")
		argIdx++
	}
	if filter.SettingType != "" {
		countQuery += fmt.Sprintf(` AND setting_type = $%d`, argIdx)
		args = append(args, filter.SettingType)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// List
	listQuery := `SELECT id, key, value, setting_type, description,
	                     to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at,
	                     to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at
	              FROM site_settings WHERE 1=1`

	listArgs := []interface{}{}
	listIdx := 1

	if filter.Key != "" {
		listQuery += fmt.Sprintf(` AND key ILIKE $%d`, listIdx)
		listArgs = append(listArgs, "%"+filter.Key+"%")
		listIdx++
	}
	if filter.SettingType != "" {
		listQuery += fmt.Sprintf(` AND setting_type = $%d`, listIdx)
		listArgs = append(listArgs, filter.SettingType)
		listIdx++
	}

	listQuery += fmt.Sprintf(` ORDER BY key ASC LIMIT $%d OFFSET $%d`, listIdx, listIdx+1)
	listArgs = append(listArgs, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.SiteSettingView
	for rows.Next() {
		v := &domain.SiteSettingView{}
		if err := rows.Scan(&v.ID, &v.Key, &v.Value, &v.SettingType, &v.Description, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}

	return items, total, nil
}
