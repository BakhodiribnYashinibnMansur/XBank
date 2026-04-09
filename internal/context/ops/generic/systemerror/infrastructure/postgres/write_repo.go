package postgres

import (
	"context"
	"encoding/json"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/domain"
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

func (r *WriteRepo) Save(ctx context.Context, e interface{}) error {
	sysErr := e.(*domain.SystemError)
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	metadataJSON, _ := json.Marshal(sysErr.Metadata)

	_, err := db.Exec(ctx,
		`INSERT INTO system_errors
		 (id, code, message, severity, category, stack_trace, request_id,
		  user_id, ip_address, path, method, metadata, resolution, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		sysErr.ID, sysErr.Code, sysErr.Message, sysErr.Severity, sysErr.Category,
		sysErr.StackTrace, sysErr.RequestID, sysErr.UserID, sysErr.IPAddress,
		sysErr.Path, sysErr.Method, metadataJSON, sysErr.Resolution,
		sysErr.CreatedAt, sysErr.UpdatedAt,
	)
	metrics.ObserveQuery("SystemErrorRepo.Save", start, err)
	return err
}

func (r *WriteRepo) Update(ctx context.Context, e interface{}) error {
	sysErr := e.(*domain.SystemError)
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`UPDATE system_errors SET resolution=$1, resolved_at=$2, resolved_by=$3, updated_at=$4
		 WHERE id=$5`,
		sysErr.Resolution, sysErr.ResolvedAt, sysErr.ResolvedBy, sysErr.UpdatedAt, sysErr.ID,
	)
	metrics.ObserveQuery("SystemErrorRepo.Update", start, err)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (interface{}, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	e := &domain.SystemError{}
	var metadataJSON []byte

	err := db.QueryRow(ctx,
		`SELECT id, code, message, severity, category, COALESCE(stack_trace,''),
		        COALESCE(request_id,''), COALESCE(user_id,''), COALESCE(ip_address,''),
		        COALESCE(path,''), COALESCE(method,''), COALESCE(metadata,'{}'),
		        resolution, resolved_at, COALESCE(resolved_by,''), created_at, updated_at
		 FROM system_errors WHERE id=$1`, id,
	).Scan(&e.ID, &e.Code, &e.Message, &e.Severity, &e.Category,
		&e.StackTrace, &e.RequestID, &e.UserID, &e.IPAddress,
		&e.Path, &e.Method, &metadataJSON,
		&e.Resolution, &e.ResolvedAt, &e.ResolvedBy, &e.CreatedAt, &e.UpdatedAt)
	metrics.ObserveQuery("SystemErrorRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &e.Metadata)
	}
	return e, nil
}
