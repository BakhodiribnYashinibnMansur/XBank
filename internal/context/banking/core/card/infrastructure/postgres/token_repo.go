package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepo struct {
	pool *pgxpool.Pool
}

func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

func (r *TokenRepo) Create(ctx context.Context, t *domain.Token) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO card_tokens (token, card_id, pan_encrypted, last_four, expires_at, created_at, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.Token, t.CardID, t.PANEncrypted, t.LastFour, t.ExpiresAt, t.CreatedAt, t.IsActive,
	)
	metrics.ObserveQuery("CardTokenRepo.Create", start, err)
	return err
}

func (r *TokenRepo) GetByToken(ctx context.Context, token string) (*domain.Token, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	t := &domain.Token{}
	err := db.QueryRow(ctx,
		`SELECT token, card_id, pan_encrypted, last_four, expires_at, created_at, is_active
		 FROM card_tokens WHERE token = $1`,
		token,
	).Scan(&t.Token, &t.CardID, &t.PANEncrypted, &t.LastFour, &t.ExpiresAt, &t.CreatedAt, &t.IsActive)
	metrics.ObserveQuery("CardTokenRepo.GetByToken", start, err)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TokenRepo) ListByCardID(ctx context.Context, cardID string) ([]*domain.Token, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT token, card_id, pan_encrypted, last_four, expires_at, created_at, is_active
		 FROM card_tokens WHERE card_id = $1 AND is_active = TRUE ORDER BY created_at DESC`,
		cardID,
	)
	if err != nil {
		metrics.ObserveQuery("CardTokenRepo.ListByCardID", start, err)
		return nil, err
	}
	defer rows.Close()

	var tokens []*domain.Token
	for rows.Next() {
		t := &domain.Token{}
		if err := rows.Scan(&t.Token, &t.CardID, &t.PANEncrypted, &t.LastFour, &t.ExpiresAt, &t.CreatedAt, &t.IsActive); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	metrics.ObserveQuery("CardTokenRepo.ListByCardID", start, nil)
	return tokens, nil
}

func (r *TokenRepo) Deactivate(ctx context.Context, token string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `UPDATE card_tokens SET is_active = FALSE WHERE token = $1`, token)
	metrics.ObserveQuery("CardTokenRepo.Deactivate", start, err)
	return err
}
