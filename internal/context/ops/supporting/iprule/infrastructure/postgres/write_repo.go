package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/domain"
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

func (r *WriteRepo) Save(ctx context.Context, rule *domain.IPRule) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO ip_rules (ip_address, rule_type, reason, expires_at, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		rule.IPAddress, rule.RuleType, rule.Reason, rule.ExpiresAt, rule.CreatedBy, rule.CreatedAt,
	).Scan(&rule.ID)
	metrics.ObserveQuery("IPRuleRepo.Save", start, err)
	if err != nil {
		return fmt.Errorf("ip_rule_repo: save: %w", err)
	}
	return nil
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.IPRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rule := &domain.IPRule{}
	err := db.QueryRow(ctx,
		`SELECT id, ip_address, rule_type, reason, expires_at, created_by, created_at
		 FROM ip_rules WHERE id = $1`, id,
	).Scan(&rule.ID, &rule.IPAddress, &rule.RuleType, &rule.Reason, &rule.ExpiresAt, &rule.CreatedBy, &rule.CreatedAt)
	metrics.ObserveQuery("IPRuleRepo.FindByID", start, err)
	if err != nil {
		return nil, domain.ErrIPRuleNotFound
	}
	return rule, nil
}

func (r *WriteRepo) FindByIP(ctx context.Context, ipAddress string) (*domain.IPRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rule := &domain.IPRule{}
	err := db.QueryRow(ctx,
		`SELECT id, ip_address, rule_type, reason, expires_at, created_by, created_at
		 FROM ip_rules WHERE ip_address = $1`, ipAddress,
	).Scan(&rule.ID, &rule.IPAddress, &rule.RuleType, &rule.Reason, &rule.ExpiresAt, &rule.CreatedBy, &rule.CreatedAt)
	metrics.ObserveQuery("IPRuleRepo.FindByIP", start, err)
	if err != nil {
		return nil, domain.ErrIPRuleNotFound
	}
	return rule, nil
}

func (r *WriteRepo) ListAll(ctx context.Context) ([]*domain.IPRule, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, ip_address, rule_type, reason, expires_at, created_by, created_at
		 FROM ip_rules ORDER BY created_at DESC`,
	)
	if err != nil {
		metrics.ObserveQuery("IPRuleRepo.ListAll", start, err)
		return nil, fmt.Errorf("ip_rule_repo: list_all: %w", err)
	}
	defer rows.Close()

	var items []*domain.IPRule
	for rows.Next() {
		rule := &domain.IPRule{}
		if err := rows.Scan(&rule.ID, &rule.IPAddress, &rule.RuleType, &rule.Reason, &rule.ExpiresAt, &rule.CreatedBy, &rule.CreatedAt); err != nil {
			metrics.ObserveQuery("IPRuleRepo.ListAll", start, err)
			return nil, fmt.Errorf("ip_rule_repo: list_all scan: %w", err)
		}
		items = append(items, rule)
	}
	metrics.ObserveQuery("IPRuleRepo.ListAll", start, nil)
	return items, nil
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM ip_rules WHERE id = $1`, id)
	metrics.ObserveQuery("IPRuleRepo.Delete", start, err)
	return err
}
