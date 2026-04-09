package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
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

func (r *WriteRepo) Upsert(ctx context.Context, s interface{}) error {
	setting := s.(*domain.UserSetting)
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO user_settings (id, user_id, key, value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		setting.ID, setting.UserID, setting.Key, setting.Value, setting.CreatedAt, setting.UpdatedAt,
	)
	metrics.ObserveQuery("UserSettingRepo.Upsert", start, err)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM user_settings WHERE id=$1`, id)
	metrics.ObserveQuery("UserSettingRepo.Delete", start, err)
	return err
}

func (r *WriteRepo) FindByUserIDAndKey(ctx context.Context, userID, key string) (interface{}, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	s := &domain.UserSetting{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, key, value, created_at, updated_at
		 FROM user_settings WHERE user_id=$1 AND key=$2`,
		userID, key,
	).Scan(&s.ID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt)
	metrics.ObserveQuery("UserSettingRepo.FindByUserIDAndKey", start, err)
	if err != nil {
		return nil, err
	}
	return s, nil
}
