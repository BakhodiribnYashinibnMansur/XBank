package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthPortAdapter implements ports.UserAuthReader using the users table.
type AuthPortAdapter struct {
	pool *pgxpool.Pool
}

func NewAuthPortAdapter(pool *pgxpool.Pool) *AuthPortAdapter {
	return &AuthPortAdapter{pool: pool}
}

func (a *AuthPortAdapter) GetByID(ctx context.Context, id string) (*ports.UserAuthView, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	v := &ports.UserAuthView{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, role, totp_secret, totp_enabled
		 FROM users WHERE id = $1`, id,
	).Scan(&v.ID, &v.Email, &v.Password, &v.Role, &v.TOTPSecret, &v.TOTPEnabled)
	metrics.ObserveQuery("UserAuthPort.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return v, nil
}

func (a *AuthPortAdapter) GetByEmail(ctx context.Context, email string) (*ports.UserAuthView, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	v := &ports.UserAuthView{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, role, totp_secret, totp_enabled
		 FROM users WHERE email = $1`, email,
	).Scan(&v.ID, &v.Email, &v.Password, &v.Role, &v.TOTPSecret, &v.TOTPEnabled)
	metrics.ObserveQuery("UserAuthPort.GetByEmail", start, err)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return v, nil
}

func (a *AuthPortAdapter) UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	_, err := db.Exec(ctx,
		`UPDATE users SET totp_secret = $1, totp_enabled = $2, totp_verified_at = $3, updated_at = NOW() WHERE id = $4`,
		totpSecret, enabled, verifiedAt, userID,
	)
	metrics.ObserveQuery("UserAuthPort.UpdateTOTP", start, err)
	return err
}
