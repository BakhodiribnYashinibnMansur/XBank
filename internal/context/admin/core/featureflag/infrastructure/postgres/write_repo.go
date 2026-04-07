package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements domain.WriteRepository for PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Save(ctx context.Context, f *domain.FeatureFlag) error {
	query := `INSERT INTO feature_flags (key, description, flag_type, default_value, active, rollout_pct, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	return r.pool.QueryRow(ctx, query,
		f.Key, f.Description, f.FlagType, f.DefaultValue, f.Active, f.RolloutPct, f.CreatedAt, f.UpdatedAt,
	).Scan(&f.ID)
}

func (r *WriteRepo) Update(ctx context.Context, f *domain.FeatureFlag) error {
	query := `UPDATE feature_flags SET description=$1, default_value=$2, active=$3, rollout_pct=$4, updated_at=$5 WHERE id=$6`
	_, err := r.pool.Exec(ctx, query, f.Description, f.DefaultValue, f.Active, f.RolloutPct, f.UpdatedAt, f.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM feature_flags WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.FeatureFlag, error) {
	f := &domain.FeatureFlag{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, description, flag_type, default_value, active, rollout_pct, created_at, updated_at
		 FROM feature_flags WHERE id = $1`, id,
	).Scan(&f.ID, &f.Key, &f.Description, &f.FlagType, &f.DefaultValue, &f.Active, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("feature_flag find by id: %w", err)
	}

	ruleGroups, _ := r.loadRuleGroups(ctx, f.ID)
	f.RuleGroups = ruleGroups
	return f, nil
}

func (r *WriteRepo) FindByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	f := &domain.FeatureFlag{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, description, flag_type, default_value, active, rollout_pct, created_at, updated_at
		 FROM feature_flags WHERE key = $1`, key,
	).Scan(&f.ID, &f.Key, &f.Description, &f.FlagType, &f.DefaultValue, &f.Active, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("feature_flag find by key: %w", err)
	}

	ruleGroups, _ := r.loadRuleGroups(ctx, f.ID)
	f.RuleGroups = ruleGroups
	return f, nil
}

func (r *WriteRepo) FindAll(ctx context.Context) ([]*domain.FeatureFlag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, key, description, flag_type, default_value, active, rollout_pct, created_at, updated_at
		 FROM feature_flags ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*domain.FeatureFlag
	for rows.Next() {
		f := &domain.FeatureFlag{}
		if err := rows.Scan(&f.ID, &f.Key, &f.Description, &f.FlagType, &f.DefaultValue, &f.Active, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		ruleGroups, _ := r.loadRuleGroups(ctx, f.ID)
		f.RuleGroups = ruleGroups
		flags = append(flags, f)
	}
	return flags, nil
}

func (r *WriteRepo) loadRuleGroups(ctx context.Context, flagID string) ([]domain.RuleGroup, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, flag_id, name, priority, value, created_at
		 FROM feature_flag_rule_groups WHERE flag_id = $1 ORDER BY priority`, flagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []domain.RuleGroup
	for rows.Next() {
		rg := domain.RuleGroup{}
		if err := rows.Scan(&rg.ID, &rg.FlagID, &rg.Name, &rg.Priority, &rg.Value, &rg.CreatedAt); err != nil {
			return nil, err
		}
		conditions, _ := r.loadConditions(ctx, rg.ID)
		rg.Conditions = conditions
		groups = append(groups, rg)
	}
	return groups, nil
}

func (r *WriteRepo) loadConditions(ctx context.Context, ruleGroupID string) ([]domain.Condition, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, rule_group_id, attribute, operator, value
		 FROM feature_flag_conditions WHERE rule_group_id = $1`, ruleGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conditions []domain.Condition
	for rows.Next() {
		c := domain.Condition{}
		if err := rows.Scan(&c.ID, &c.RuleGroupID, &c.Attribute, &c.Operator, &c.Value); err != nil {
			return nil, err
		}
		conditions = append(conditions, c)
	}
	return conditions, nil
}
