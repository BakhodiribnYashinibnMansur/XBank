package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/domain"
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

func (r *WriteRepo) Save(ctx context.Context, rule *domain.RateLimitRule) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO rate_limit_rules (key, max_requests, window_seconds, description, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		rule.Key, rule.MaxRequests, rule.WindowSeconds, rule.Description, rule.Enabled, rule.CreatedAt, rule.UpdatedAt,
	).Scan(&rule.ID)
	metrics.ObserveQuery("RateLimitRepo.Save", start, err)
	if err != nil {
		return fmt.Errorf("rate_limit_repo: save: %w", err)
	}
	return nil
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.RateLimitRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rule := &domain.RateLimitRule{}
	err := db.QueryRow(ctx,
		`SELECT id, key, max_requests, window_seconds, description, enabled, created_at, updated_at
		 FROM rate_limit_rules WHERE id = $1`, id,
	).Scan(&rule.ID, &rule.Key, &rule.MaxRequests, &rule.WindowSeconds, &rule.Description, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
	metrics.ObserveQuery("RateLimitRepo.FindByID", start, err)
	if err != nil {
		return nil, domain.ErrRateLimitNotFound
	}
	return rule, nil
}

func (r *WriteRepo) FindByKey(ctx context.Context, key string) (*domain.RateLimitRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rule := &domain.RateLimitRule{}
	err := db.QueryRow(ctx,
		`SELECT id, key, max_requests, window_seconds, description, enabled, created_at, updated_at
		 FROM rate_limit_rules WHERE key = $1`, key,
	).Scan(&rule.ID, &rule.Key, &rule.MaxRequests, &rule.WindowSeconds, &rule.Description, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
	metrics.ObserveQuery("RateLimitRepo.FindByKey", start, err)
	if err != nil {
		return nil, domain.ErrRateLimitNotFound
	}
	return rule, nil
}

func (r *WriteRepo) FindAll(ctx context.Context) ([]*domain.RateLimitRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, key, max_requests, window_seconds, description, enabled, created_at, updated_at
		 FROM rate_limit_rules ORDER BY created_at DESC`,
	)
	if err != nil {
		metrics.ObserveQuery("RateLimitRepo.FindAll", start, err)
		return nil, fmt.Errorf("rate_limit_repo: find_all: %w", err)
	}
	defer rows.Close()

	var items []*domain.RateLimitRule
	for rows.Next() {
		rule := &domain.RateLimitRule{}
		if err := rows.Scan(&rule.ID, &rule.Key, &rule.MaxRequests, &rule.WindowSeconds, &rule.Description, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			metrics.ObserveQuery("RateLimitRepo.FindAll", start, err)
			return nil, fmt.Errorf("rate_limit_repo: find_all scan: %w", err)
		}
		items = append(items, rule)
	}
	metrics.ObserveQuery("RateLimitRepo.FindAll", start, nil)
	return items, nil
}

func (r *WriteRepo) Update(ctx context.Context, rule *domain.RateLimitRule) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE rate_limit_rules SET max_requests = $1, window_seconds = $2, description = $3, enabled = $4, updated_at = $5 WHERE id = $6`,
		rule.MaxRequests, rule.WindowSeconds, rule.Description, rule.Enabled, rule.UpdatedAt, rule.ID,
	)
	metrics.ObserveQuery("RateLimitRepo.Update", start, err)
	if err != nil {
		return fmt.Errorf("rate_limit_repo: update: %w", err)
	}
	return nil
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM rate_limit_rules WHERE id = $1`, id)
	metrics.ObserveQuery("RateLimitRepo.Delete", start, err)
	return err
}
