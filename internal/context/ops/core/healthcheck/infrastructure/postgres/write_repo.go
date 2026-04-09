package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements healthcheck.Repository using PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

// NewWriteRepo creates a new healthcheck postgres repository.
func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Save(ctx context.Context, record *domain.HealthRecord) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	err := db.QueryRow(ctx,
		`INSERT INTO health_records (status, components, checked_at)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		record.Status, record.Components, record.CheckedAt,
	).Scan(&record.ID)
	metrics.ObserveQuery("HealthcheckRepo.Save", start, err)
	if err != nil {
		return fmt.Errorf("healthcheck_repo: save: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetLatest(ctx context.Context) (*domain.HealthRecord, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	record := &domain.HealthRecord{}
	err := db.QueryRow(ctx,
		`SELECT id, status, components, checked_at
		 FROM health_records ORDER BY checked_at DESC LIMIT 1`,
	).Scan(&record.ID, &record.Status, &record.Components, &record.CheckedAt)
	metrics.ObserveQuery("HealthcheckRepo.GetLatest", start, err)
	if err != nil {
		return nil, domain.ErrCheckNotFound
	}
	return record, nil
}

func (r *WriteRepo) ListHistory(ctx context.Context, limit, offset int) ([]*domain.HealthRecord, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT id, status, components, checked_at
		 FROM health_records ORDER BY checked_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("HealthcheckRepo.ListHistory", start, err)
		return nil, fmt.Errorf("healthcheck_repo: list: %w", err)
	}
	defer rows.Close()

	var records []*domain.HealthRecord
	for rows.Next() {
		record := &domain.HealthRecord{}
		if err := rows.Scan(&record.ID, &record.Status, &record.Components, &record.CheckedAt); err != nil {
			metrics.ObserveQuery("HealthcheckRepo.ListHistory", start, err)
			return nil, fmt.Errorf("healthcheck_repo: list scan: %w", err)
		}
		records = append(records, record)
	}
	metrics.ObserveQuery("HealthcheckRepo.ListHistory", start, nil)
	return records, nil
}

func (r *WriteRepo) CountHistory(ctx context.Context) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM health_records`).Scan(&count)
	metrics.ObserveQuery("HealthcheckRepo.CountHistory", start, err)
	return count, err
}
