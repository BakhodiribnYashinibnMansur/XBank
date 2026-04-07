package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/domain"
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

func (r *WriteRepo) Upsert(ctx context.Context, fp *domain.Fingerprint) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	fp.LastUsedAt = time.Now()
	err := db.QueryRow(ctx,
		`INSERT INTO device_fingerprints (user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, device_hash) DO UPDATE
		 SET last_used_at = $6, ip_address = $4
		 RETURNING id`,
		fp.UserID, fp.DeviceHash, fp.DeviceName, fp.IPAddress, fp.Trusted, fp.LastUsedAt, time.Now(),
	).Scan(&fp.ID)
	metrics.ObserveQuery("DeviceRepo.Upsert", start, err)
	return err
}

func (r *WriteRepo) GetByUserAndDevice(ctx context.Context, userID, deviceHash string) (*domain.Fingerprint, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	fp := &domain.Fingerprint{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at
		 FROM device_fingerprints WHERE user_id = $1 AND device_hash = $2`,
		userID, deviceHash,
	).Scan(&fp.ID, &fp.UserID, &fp.DeviceHash, &fp.DeviceName, &fp.IPAddress, &fp.Trusted, &fp.LastUsedAt, &fp.CreatedAt)
	metrics.ObserveQuery("DeviceRepo.GetByUserAndDevice", start, err)
	if err != nil {
		return nil, nil
	}
	return fp, nil
}

func (r *WriteRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.Fingerprint, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at
		 FROM device_fingerprints WHERE user_id = $1 ORDER BY last_used_at DESC`, userID)
	if err != nil {
		metrics.ObserveQuery("DeviceRepo.ListByUserID", start, err)
		return nil, err
	}
	defer rows.Close()

	var fps []*domain.Fingerprint
	for rows.Next() {
		fp := &domain.Fingerprint{}
		if err := rows.Scan(&fp.ID, &fp.UserID, &fp.DeviceHash, &fp.DeviceName, &fp.IPAddress, &fp.Trusted, &fp.LastUsedAt, &fp.CreatedAt); err != nil {
			metrics.ObserveQuery("DeviceRepo.ListByUserID", start, err)
			return nil, err
		}
		fps = append(fps, fp)
	}
	metrics.ObserveQuery("DeviceRepo.ListByUserID", start, nil)
	return fps, nil
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM device_fingerprints WHERE id = $1`, id)
	metrics.ObserveQuery("DeviceRepo.Delete", start, err)
	return err
}
