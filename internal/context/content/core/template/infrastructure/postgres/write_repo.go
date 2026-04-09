package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements template.Repository using PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

// NewWriteRepo creates a new template postgres repository.
func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Create(ctx context.Context, t *domain.Template) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt

	err := db.QueryRow(ctx,
		`INSERT INTO templates (slug, channel, subject, body, locale, status, version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		t.Slug, t.Channel, t.Subject, t.Body, t.Locale, t.Status, t.Version, t.CreatedAt, t.UpdatedAt,
	).Scan(&t.ID)
	metrics.ObserveQuery("TemplateRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("template_repo: create: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	t := &domain.Template{}
	err := db.QueryRow(ctx,
		`SELECT id, slug, channel, subject, body, locale, status, version, created_at, updated_at
		 FROM templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.Slug, &t.Channel, &t.Subject, &t.Body, &t.Locale, &t.Status, &t.Version, &t.CreatedAt, &t.UpdatedAt)
	metrics.ObserveQuery("TemplateRepo.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrTemplateNotFound
	}
	return t, nil
}

func (r *WriteRepo) GetBySlugAndLocale(ctx context.Context, slug, locale string) (*domain.Template, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	t := &domain.Template{}
	err := db.QueryRow(ctx,
		`SELECT id, slug, channel, subject, body, locale, status, version, created_at, updated_at
		 FROM templates WHERE slug = $1 AND locale = $2 AND status = 'ACTIVE'`, slug, locale,
	).Scan(&t.ID, &t.Slug, &t.Channel, &t.Subject, &t.Body, &t.Locale, &t.Status, &t.Version, &t.CreatedAt, &t.UpdatedAt)
	metrics.ObserveQuery("TemplateRepo.GetBySlugAndLocale", start, err)
	if err != nil {
		return nil, domain.ErrTemplateNotFound
	}
	return t, nil
}

func (r *WriteRepo) ListByChannel(ctx context.Context, channel string, limit, offset int) ([]*domain.Template, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	query := `SELECT id, slug, channel, subject, body, locale, status, version, created_at, updated_at
		 FROM templates WHERE ($1 = '' OR channel = $1) ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`

	rows, err := db.Query(ctx, query, channel, limit, offset)
	if err != nil {
		metrics.ObserveQuery("TemplateRepo.ListByChannel", start, err)
		return nil, fmt.Errorf("template_repo: list: %w", err)
	}
	defer rows.Close()

	var templates []*domain.Template
	for rows.Next() {
		t := &domain.Template{}
		if err := rows.Scan(&t.ID, &t.Slug, &t.Channel, &t.Subject, &t.Body, &t.Locale, &t.Status, &t.Version, &t.CreatedAt, &t.UpdatedAt); err != nil {
			metrics.ObserveQuery("TemplateRepo.ListByChannel", start, err)
			return nil, fmt.Errorf("template_repo: list scan: %w", err)
		}
		templates = append(templates, t)
	}
	metrics.ObserveQuery("TemplateRepo.ListByChannel", start, nil)
	return templates, nil
}

func (r *WriteRepo) CountByChannel(ctx context.Context, channel string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM templates WHERE ($1 = '' OR channel = $1)`, channel,
	).Scan(&count)
	metrics.ObserveQuery("TemplateRepo.CountByChannel", start, err)
	return count, err
}

func (r *WriteRepo) Update(ctx context.Context, t *domain.Template) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	t.UpdatedAt = time.Now()

	_, err := db.Exec(ctx,
		`UPDATE templates SET subject = $1, body = $2, status = $3, version = $4, updated_at = $5
		 WHERE id = $6`,
		t.Subject, t.Body, t.Status, t.Version, t.UpdatedAt, t.ID,
	)
	metrics.ObserveQuery("TemplateRepo.Update", start, err)
	if err != nil {
		return fmt.Errorf("template_repo: update: %w", err)
	}
	return nil
}
