package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Save(ctx context.Context, i *domain.Integration) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO integrations (name, base_url, api_key, status, webhook_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		i.Name, i.BaseURL, i.APIKey, i.Status, i.WebhookURL, i.CreatedAt, i.UpdatedAt,
	).Scan(&i.ID)
	metrics.ObserveQuery("IntegrationRepo.Save", start, err)
	if err != nil {
		return fmt.Errorf("integration_repo: save: %w", err)
	}
	return nil
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.Integration, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	i := &domain.Integration{}
	err := db.QueryRow(ctx,
		`SELECT id, name, base_url, api_key, status, webhook_url, created_at, updated_at
		 FROM integrations WHERE id = $1`, id,
	).Scan(&i.ID, &i.Name, &i.BaseURL, &i.APIKey, &i.Status, &i.WebhookURL, &i.CreatedAt, &i.UpdatedAt)
	metrics.ObserveQuery("IntegrationRepo.FindByID", start, err)
	if err != nil {
		return nil, domain.ErrIntegrationNotFound
	}
	return i, nil
}

func (r *WriteRepo) FindByName(ctx context.Context, name string) (*domain.Integration, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	i := &domain.Integration{}
	err := db.QueryRow(ctx,
		`SELECT id, name, base_url, api_key, status, webhook_url, created_at, updated_at
		 FROM integrations WHERE name = $1`, name,
	).Scan(&i.ID, &i.Name, &i.BaseURL, &i.APIKey, &i.Status, &i.WebhookURL, &i.CreatedAt, &i.UpdatedAt)
	metrics.ObserveQuery("IntegrationRepo.FindByName", start, err)
	if err != nil {
		return nil, domain.ErrIntegrationNotFound
	}
	return i, nil
}

func (r *WriteRepo) ListAll(ctx context.Context) ([]*domain.Integration, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, name, base_url, api_key, status, webhook_url, created_at, updated_at
		 FROM integrations ORDER BY created_at DESC`,
	)
	if err != nil {
		metrics.ObserveQuery("IntegrationRepo.ListAll", start, err)
		return nil, fmt.Errorf("integration_repo: list_all: %w", err)
	}
	defer rows.Close()

	var items []*domain.Integration
	for rows.Next() {
		i := &domain.Integration{}
		if err := rows.Scan(&i.ID, &i.Name, &i.BaseURL, &i.APIKey, &i.Status, &i.WebhookURL, &i.CreatedAt, &i.UpdatedAt); err != nil {
			metrics.ObserveQuery("IntegrationRepo.ListAll", start, err)
			return nil, fmt.Errorf("integration_repo: list_all scan: %w", err)
		}
		items = append(items, i)
	}
	metrics.ObserveQuery("IntegrationRepo.ListAll", start, nil)
	return items, nil
}

func (r *WriteRepo) Update(ctx context.Context, i *domain.Integration) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE integrations SET base_url = $1, api_key = $2, status = $3, webhook_url = $4, updated_at = $5 WHERE id = $6`,
		i.BaseURL, i.APIKey, i.Status, i.WebhookURL, i.UpdatedAt, i.ID,
	)
	metrics.ObserveQuery("IntegrationRepo.Update", start, err)
	if err != nil {
		return fmt.Errorf("integration_repo: update: %w", err)
	}
	return nil
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	metrics.ObserveQuery("IntegrationRepo.Delete", start, err)
	return err
}
