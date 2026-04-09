package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/domain"
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

func (r *WriteRepo) Save(ctx context.Context, m *domain.AppMetric) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	labelsJSON, err := json.Marshal(m.Labels)
	if err != nil {
		return fmt.Errorf("metric_repo: marshal labels: %w", err)
	}

	err = db.QueryRow(ctx,
		`INSERT INTO app_metrics (name, value, labels, collected_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		m.Name, m.Value, labelsJSON, m.CollectedAt,
	).Scan(&m.ID)
	metrics.ObserveQuery("MetricRepo.Save", start, err)
	if err != nil {
		return fmt.Errorf("metric_repo: save: %w", err)
	}
	return nil
}

func (r *WriteRepo) FindByName(ctx context.Context, name string) ([]*domain.AppMetric, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, name, value, labels, collected_at
		 FROM app_metrics WHERE name = $1 ORDER BY collected_at DESC`, name,
	)
	if err != nil {
		metrics.ObserveQuery("MetricRepo.FindByName", start, err)
		return nil, fmt.Errorf("metric_repo: find_by_name: %w", err)
	}
	defer rows.Close()

	return scanMetrics(rows, "MetricRepo.FindByName", start)
}

func (r *WriteRepo) ListRecent(ctx context.Context, limit int) ([]*domain.AppMetric, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, name, value, labels, collected_at
		 FROM app_metrics ORDER BY collected_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		metrics.ObserveQuery("MetricRepo.ListRecent", start, err)
		return nil, fmt.Errorf("metric_repo: list_recent: %w", err)
	}
	defer rows.Close()

	return scanMetrics(rows, "MetricRepo.ListRecent", start)
}

type scannable interface {
	Next() bool
	Scan(dest ...interface{}) error
}

func scanMetrics(rows scannable, operation string, start time.Time) ([]*domain.AppMetric, error) {
	var items []*domain.AppMetric
	for rows.Next() {
		m := &domain.AppMetric{}
		var labelsJSON []byte
		if err := rows.Scan(&m.ID, &m.Name, &m.Value, &labelsJSON, &m.CollectedAt); err != nil {
			metrics.ObserveQuery(operation, start, err)
			return nil, fmt.Errorf("metric_repo: scan: %w", err)
		}
		if len(labelsJSON) > 0 {
			if err := json.Unmarshal(labelsJSON, &m.Labels); err != nil {
				return nil, fmt.Errorf("metric_repo: unmarshal labels: %w", err)
			}
		}
		if m.Labels == nil {
			m.Labels = make(map[string]string)
		}
		items = append(items, m)
	}
	metrics.ObserveQuery(operation, start, nil)
	return items, nil
}
