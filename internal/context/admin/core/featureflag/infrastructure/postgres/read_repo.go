package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReadRepo implements repository.ReadRepository for feature flags.
type ReadRepo struct {
	pool *pgxpool.Pool
}

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo {
	return &ReadRepo{pool: pool}
}

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*repository.FeatureFlagView, error) {
	v := &repository.FeatureFlagView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, key, description, flag_type, default_value, active, rollout_pct,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		        to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM feature_flags WHERE id = $1`, id,
	).Scan(&v.ID, &v.Key, &v.Description, &v.FlagType, &v.DefaultValue, &v.Active, &v.RolloutPct, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("feature_flag read: %w", err)
	}

	v.RuleGroups, _ = r.loadRuleGroupViews(ctx, v.ID)
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter repository.FeatureFlagFilter) ([]*repository.FeatureFlagView, int64, error) {
	countQuery := `SELECT COUNT(*) FROM feature_flags WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if filter.Key != "" {
		countQuery += fmt.Sprintf(` AND key ILIKE $%d`, idx)
		args = append(args, "%"+filter.Key+"%")
		idx++
	}
	if filter.Active != nil {
		countQuery += fmt.Sprintf(` AND active = $%d`, idx)
		args = append(args, *filter.Active)
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT id, key, description, flag_type, default_value, active, rollout_pct,
	                     to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
	                     to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	              FROM feature_flags WHERE 1=1`
	listArgs := []interface{}{}
	listIdx := 1

	if filter.Key != "" {
		listQuery += fmt.Sprintf(` AND key ILIKE $%d`, listIdx)
		listArgs = append(listArgs, "%"+filter.Key+"%")
		listIdx++
	}
	if filter.Active != nil {
		listQuery += fmt.Sprintf(` AND active = $%d`, listIdx)
		listArgs = append(listArgs, *filter.Active)
		listIdx++
	}

	listQuery += fmt.Sprintf(` ORDER BY key ASC LIMIT $%d OFFSET $%d`, listIdx, listIdx+1)
	listArgs = append(listArgs, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*repository.FeatureFlagView
	for rows.Next() {
		v := &repository.FeatureFlagView{}
		if err := rows.Scan(&v.ID, &v.Key, &v.Description, &v.FlagType, &v.DefaultValue, &v.Active, &v.RolloutPct, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}

	return items, total, nil
}

func (r *ReadRepo) loadRuleGroupViews(ctx context.Context, flagID string) ([]repository.RuleGroupView, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, priority, value FROM feature_flag_rule_groups WHERE flag_id = $1 ORDER BY priority`, flagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []repository.RuleGroupView
	for rows.Next() {
		rg := repository.RuleGroupView{}
		if err := rows.Scan(&rg.ID, &rg.Name, &rg.Priority, &rg.Value); err != nil {
			return nil, err
		}
		rg.Conditions, _ = r.loadConditionViews(ctx, rg.ID)
		groups = append(groups, rg)
	}
	return groups, nil
}

func (r *ReadRepo) loadConditionViews(ctx context.Context, ruleGroupID string) ([]repository.ConditionView, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, attribute, operator, value FROM feature_flag_conditions WHERE rule_group_id = $1`, ruleGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conditions []repository.ConditionView
	for rows.Next() {
		c := repository.ConditionView{}
		if err := rows.Scan(&c.ID, &c.Attribute, &c.Operator, &c.Value); err != nil {
			return nil, err
		}
		conditions = append(conditions, c)
	}
	return conditions, nil
}
