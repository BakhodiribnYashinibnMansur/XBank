package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
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

func (r *WriteRepo) Create(ctx context.Context, u *domain.User) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO users (email, password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := db.QueryRow(ctx, query,
		u.Email, u.Password, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
	metrics.ObserveQuery("UserRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	u := &domain.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.TOTPSecret, &u.TOTPEnabled, &u.TOTPVerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	metrics.ObserveQuery("UserRepo.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *WriteRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	u := &domain.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, totp_secret, totp_enabled, totp_verified_at, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.TOTPSecret, &u.TOTPEnabled, &u.TOTPVerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	metrics.ObserveQuery("UserRepo.GetByEmail", start, err)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *WriteRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	metrics.ObserveQuery("UserRepo.ExistsByEmail", start, err)
	return exists, err
}

func (r *WriteRepo) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`, hashedPassword, userID)
	metrics.ObserveQuery("UserRepo.UpdatePassword", start, err)
	return err
}

func (r *WriteRepo) UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE users SET totp_secret = $1, totp_enabled = $2, totp_verified_at = $3, updated_at = NOW() WHERE id = $4`,
		totpSecret, enabled, verifiedAt, userID,
	)
	metrics.ObserveQuery("UserRepo.UpdateTOTP", start, err)
	return err
}

func (r *WriteRepo) Anonymize(ctx context.Context, userID string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE users SET email = 'deleted_' || id, password = '', first_name = '[DELETED]', last_name = '[DELETED]', updated_at = NOW() WHERE id = $1`,
		userID,
	)
	metrics.ObserveQuery("UserRepo.Anonymize", start, err)
	return err
}
