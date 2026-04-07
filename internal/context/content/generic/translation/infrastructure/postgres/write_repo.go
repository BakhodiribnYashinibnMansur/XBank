package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Save(ctx context.Context, t *domain.Translation) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO translations (key, language, value, "group", created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		t.Key, t.Language, t.Value, t.Group, t.CreatedAt, t.UpdatedAt,
	).Scan(&t.ID)
}

func (r *WriteRepo) Update(ctx context.Context, t *domain.Translation) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE translations SET value = $1, updated_at = $2 WHERE id = $3`,
		t.Value, t.UpdatedAt, t.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM translations WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.Translation, error) {
	t := &domain.Translation{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, language, value, "group", created_at, updated_at FROM translations WHERE id = $1`, id,
	).Scan(&t.ID, &t.Key, &t.Language, &t.Value, &t.Group, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("translation find by id: %w", err)
	}
	return t, nil
}

func (r *WriteRepo) FindByKeyAndLanguage(ctx context.Context, key string, lang domain.Language) (*domain.Translation, error) {
	t := &domain.Translation{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, language, value, "group", created_at, updated_at
		 FROM translations WHERE key = $1 AND language = $2`, key, lang,
	).Scan(&t.ID, &t.Key, &t.Language, &t.Value, &t.Group, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("translation find by key+lang: %w", err)
	}
	return t, nil
}
