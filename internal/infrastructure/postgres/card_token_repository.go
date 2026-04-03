package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardTokenRepository struct {
	pool *pgxpool.Pool
}

func NewCardTokenRepository(pool *pgxpool.Pool) *CardTokenRepository {
	return &CardTokenRepository{pool: pool}
}

func (r *CardTokenRepository) Create(ctx context.Context, t *card.Token) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO card_tokens (token, card_id, pan_encrypted, last_four, expires_at, created_at, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.Token, t.CardID, t.PANEncrypted, t.LastFour, t.ExpiresAt, t.CreatedAt, t.IsActive,
	)
	metrics.ObserveQuery("CardTokenRepo.Create", start, err)
	return err
}

func (r *CardTokenRepository) GetByToken(ctx context.Context, token string) (*card.Token, error) {
	start := time.Now()
	t := &card.Token{}
	err := r.pool.QueryRow(ctx,
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

func (r *CardTokenRepository) ListByCardID(ctx context.Context, cardID string) ([]*card.Token, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT token, card_id, pan_encrypted, last_four, expires_at, created_at, is_active
		 FROM card_tokens WHERE card_id = $1 AND is_active = TRUE ORDER BY created_at DESC`,
		cardID,
	)
	if err != nil {
		metrics.ObserveQuery("CardTokenRepo.ListByCardID", start, err)
		return nil, err
	}
	defer rows.Close()

	var tokens []*card.Token
	for rows.Next() {
		t := &card.Token{}
		if err := rows.Scan(&t.Token, &t.CardID, &t.PANEncrypted, &t.LastFour, &t.ExpiresAt, &t.CreatedAt, &t.IsActive); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	metrics.ObserveQuery("CardTokenRepo.ListByCardID", start, nil)
	return tokens, nil
}

func (r *CardTokenRepository) Deactivate(ctx context.Context, token string) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx, `UPDATE card_tokens SET is_active = FALSE WHERE token = $1`, token)
	metrics.ObserveQuery("CardTokenRepo.Deactivate", start, err)
	return err
}
