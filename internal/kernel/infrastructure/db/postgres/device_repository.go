package postgres

import (
	"context"
	"time"

	device "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{pool: pool}
}

func (r *DeviceRepository) Upsert(ctx context.Context, fp *device.Fingerprint) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
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

func (r *DeviceRepository) GetByUserAndDevice(ctx context.Context, userID, deviceHash string) (*device.Fingerprint, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	fp := &device.Fingerprint{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at
		 FROM device_fingerprints WHERE user_id = $1 AND device_hash = $2`,
		userID, deviceHash,
	).Scan(&fp.ID, &fp.UserID, &fp.DeviceHash, &fp.DeviceName, &fp.IPAddress, &fp.Trusted, &fp.LastUsedAt, &fp.CreatedAt)
	metrics.ObserveQuery("DeviceRepo.GetByUserAndDevice", start, err)
	if err != nil {
		return nil, nil // not found = new device
	}
	return fp, nil
}

func (r *DeviceRepository) ListByUserID(ctx context.Context, userID string) ([]*device.Fingerprint, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at
		 FROM device_fingerprints WHERE user_id = $1 ORDER BY last_used_at DESC`, userID)
	if err != nil {
		metrics.ObserveQuery("DeviceRepo.ListByUserID", start, err)
		return nil, err
	}
	defer rows.Close()

	var fps []*device.Fingerprint
	for rows.Next() {
		fp := &device.Fingerprint{}
		if err := rows.Scan(&fp.ID, &fp.UserID, &fp.DeviceHash, &fp.DeviceName, &fp.IPAddress, &fp.Trusted, &fp.LastUsedAt, &fp.CreatedAt); err != nil {
			metrics.ObserveQuery("DeviceRepo.ListByUserID", start, err)
			return nil, err
		}
		fps = append(fps, fp)
	}
	metrics.ObserveQuery("DeviceRepo.ListByUserID", start, nil)
	return fps, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM device_fingerprints WHERE id = $1`, id)
	metrics.ObserveQuery("DeviceRepo.Delete", start, err)
	return err
}
