package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
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

func (r *WriteRepo) Save(ctx context.Context, e *domain.DataExport) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO data_exports (id, user_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		e.ID, e.UserID, e.Status, e.CreatedAt, e.UpdatedAt,
	)
	metrics.ObserveQuery("DataExportRepo.Save", start, err)
	return err
}

func (r *WriteRepo) Update(ctx context.Context, e *domain.DataExport) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE data_exports SET status=$1, file_url=$2, error_msg=$3, updated_at=$4
		 WHERE id=$5`,
		e.Status, e.FileURL, e.ErrorMsg, e.UpdatedAt, e.ID,
	)
	metrics.ObserveQuery("DataExportRepo.Update", start, err)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.DataExport, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	e := &domain.DataExport{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, status, file_url, error_msg, created_at, updated_at
		 FROM data_exports WHERE id=$1`, id,
	).Scan(&e.ID, &e.UserID, &e.Status, &e.FileURL, &e.ErrorMsg, &e.CreatedAt, &e.UpdatedAt)
	metrics.ObserveQuery("DataExportRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}
	return e, nil
}
