package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
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

func (r *WriteRepo) Save(ctx context.Context, f interface{}) error {
	file := f.(*domain.File)
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO files (id, name, original_name, mime_type, size, path, url, uploaded_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		file.ID, file.Name, file.OriginalName, file.MimeType, file.Size,
		file.Path, file.URL, file.UploadedBy, file.CreatedAt, file.UpdatedAt,
	)
	metrics.ObserveQuery("FileRepo.Save", start, err)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	metrics.ObserveQuery("FileRepo.Delete", start, err)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (interface{}, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	f := &domain.File{}
	err := db.QueryRow(ctx,
		`SELECT id, name, original_name, mime_type, size, path, url, uploaded_by, created_at, updated_at
		 FROM files WHERE id=$1`, id,
	).Scan(&f.ID, &f.Name, &f.OriginalName, &f.MimeType, &f.Size,
		&f.Path, &f.URL, &f.UploadedBy, &f.CreatedAt, &f.UpdatedAt)
	metrics.ObserveQuery("FileRepo.FindByID", start, err)
	if err != nil {
		return nil, err
	}
	return f, nil
}
