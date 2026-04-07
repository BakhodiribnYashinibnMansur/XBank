package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements domain.WriteRepository for PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Save(ctx context.Context, s *domain.SiteSetting) error {
	query := `INSERT INTO site_settings (key, value, setting_type, description, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.pool.QueryRow(ctx, query,
		s.Key, s.Value, s.SettingType, s.Description, s.CreatedAt, s.UpdatedAt,
	).Scan(&s.ID)
}

func (r *WriteRepo) Update(ctx context.Context, s *domain.SiteSetting) error {
	query := `UPDATE site_settings SET value = $1, description = $2, updated_at = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, s.Value, s.Description, s.UpdatedAt, s.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM site_settings WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.SiteSetting, error) {
	s := &domain.SiteSetting{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, value, setting_type, description, created_at, updated_at
		 FROM site_settings WHERE id = $1`, id,
	).Scan(&s.ID, &s.Key, &s.Value, &s.SettingType, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("site_setting find by id: %w", err)
	}
	return s, nil
}

func (r *WriteRepo) FindByKey(ctx context.Context, key string) (*domain.SiteSetting, error) {
	s := &domain.SiteSetting{}
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, value, setting_type, description, created_at, updated_at
		 FROM site_settings WHERE key = $1`, key,
	).Scan(&s.ID, &s.Key, &s.Value, &s.SettingType, &s.Description, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("site_setting find by key: %w", err)
	}
	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt
	return s, nil
}
