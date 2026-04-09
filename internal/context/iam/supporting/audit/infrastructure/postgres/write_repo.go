package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/domain"
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

// ── Audit Logs ──────────────────────────────────────────────────────

func (r *WriteRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	attrs, err := json.Marshal(log.Attributes)
	if err != nil {
		return fmt.Errorf("audit_repo: marshal attributes: %w", err)
	}

	err = db.QueryRow(ctx,
		`INSERT INTO audit_logs (aggregate_type, aggregate_id, action, actor_id, attributes, ip_address, user_agent, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		log.AggregateType, log.AggregateID, log.Action, log.ActorID, attrs, log.IPAddress, log.UserAgent, log.CreatedAt,
	).Scan(&log.ID)
	metrics.ObserveQuery("AuditRepo.CreateAuditLog", start, err)
	if err != nil {
		return fmt.Errorf("audit_repo: create_audit_log: %w", err)
	}
	return nil
}

func (r *WriteRepo) ListAuditLogs(ctx context.Context, filter domain.AuditFilter) ([]*domain.AuditLog, int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	where, args := buildAuditWhere(filter)

	// Count
	var total int64
	countQuery := "SELECT COUNT(*) FROM audit_logs" + where
	err := db.QueryRow(ctx, countQuery, args...).Scan(&total)
	metrics.ObserveQuery("AuditRepo.ListAuditLogs.Count", start, err)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_repo: count_audit_logs: %w", err)
	}

	// Rows
	dataQuery := fmt.Sprintf(
		`SELECT id, aggregate_type, aggregate_id, action, actor_id, attributes, ip_address, user_agent, created_at
		 FROM audit_logs%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)+1, len(args)+2,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.Query(ctx, dataQuery, args...)
	metrics.ObserveQuery("AuditRepo.ListAuditLogs", start, err)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_repo: list_audit_logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		l := &domain.AuditLog{}
		var attrsJSON []byte
		if err := rows.Scan(&l.ID, &l.AggregateType, &l.AggregateID, &l.Action, &l.ActorID, &attrsJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("audit_repo: scan_audit_log: %w", err)
		}
		if err := json.Unmarshal(attrsJSON, &l.Attributes); err != nil {
			l.Attributes = map[string]any{}
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

func buildAuditWhere(f domain.AuditFilter) (string, []any) {
	var conditions []string
	var args []any
	n := 1

	if f.AggregateType != "" {
		conditions = append(conditions, fmt.Sprintf("aggregate_type = $%d", n))
		args = append(args, f.AggregateType)
		n++
	}
	if f.AggregateID != "" {
		conditions = append(conditions, fmt.Sprintf("aggregate_id = $%d", n))
		args = append(args, f.AggregateID)
		n++
	}
	if f.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", n))
		args = append(args, f.Action)
		n++
	}
	if f.ActorID != "" {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", n))
		args = append(args, f.ActorID)
		n++
	}
	if f.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", n))
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", n))
		args = append(args, *f.To)
		n++
	}

	if len(conditions) == 0 {
		return "", nil
	}

	where := " WHERE "
	for i, c := range conditions {
		if i > 0 {
			where += " AND "
		}
		where += c
	}
	return where, args
}

// ── Endpoint History ────────────────────────────────────────────────

func (r *WriteRepo) CreateEndpointHistory(ctx context.Context, h *domain.EndpointHistory) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	err := db.QueryRow(ctx,
		`INSERT INTO endpoint_history (method, path, status_code, user_id, ip_address, duration_ms, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		h.Method, h.Path, h.StatusCode, h.UserID, h.IPAddress, h.DurationMs, h.CreatedAt,
	).Scan(&h.ID)
	metrics.ObserveQuery("AuditRepo.CreateEndpointHistory", start, err)
	if err != nil {
		return fmt.Errorf("audit_repo: create_endpoint_history: %w", err)
	}
	return nil
}

func (r *WriteRepo) ListEndpointHistory(ctx context.Context, filter domain.EndpointFilter) ([]*domain.EndpointHistory, int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	where, args := buildEndpointWhere(filter)

	var total int64
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM endpoint_history"+where, args...).Scan(&total)
	metrics.ObserveQuery("AuditRepo.ListEndpointHistory.Count", start, err)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_repo: count_endpoint_history: %w", err)
	}

	dataQuery := fmt.Sprintf(
		`SELECT id, method, path, status_code, user_id, ip_address, duration_ms, created_at
		 FROM endpoint_history%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)+1, len(args)+2,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.Query(ctx, dataQuery, args...)
	metrics.ObserveQuery("AuditRepo.ListEndpointHistory", start, err)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_repo: list_endpoint_history: %w", err)
	}
	defer rows.Close()

	var entries []*domain.EndpointHistory
	for rows.Next() {
		h := &domain.EndpointHistory{}
		if err := rows.Scan(&h.ID, &h.Method, &h.Path, &h.StatusCode, &h.UserID, &h.IPAddress, &h.DurationMs, &h.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("audit_repo: scan_endpoint_history: %w", err)
		}
		entries = append(entries, h)
	}
	return entries, total, nil
}

func buildEndpointWhere(f domain.EndpointFilter) (string, []any) {
	var conditions []string
	var args []any
	n := 1

	if f.Method != "" {
		conditions = append(conditions, fmt.Sprintf("method = $%d", n))
		args = append(args, f.Method)
		n++
	}
	if f.Path != "" {
		conditions = append(conditions, fmt.Sprintf("path = $%d", n))
		args = append(args, f.Path)
		n++
	}
	if f.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", n))
		args = append(args, f.UserID)
		n++
	}
	if f.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", n))
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", n))
		args = append(args, *f.To)
		n++
	}

	if len(conditions) == 0 {
		return "", nil
	}

	where := " WHERE "
	for i, c := range conditions {
		if i > 0 {
			where += " AND "
		}
		where += c
	}
	return where, args
}
