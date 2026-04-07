package postgres

import (
	"context"
	"time"

	card "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardHoldRepository struct {
	pool *pgxpool.Pool
}

func NewCardHoldRepository(pool *pgxpool.Pool) *CardHoldRepository {
	return &CardHoldRepository{pool: pool}
}

func (r *CardHoldRepository) Create(ctx context.Context, h *card.Hold) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO card_holds (id, card_id, account_id, merchant, amount, currency, status, held_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		h.ID, h.CardID, h.AccountID, h.Merchant, h.Amount, h.Currency, h.Status, h.HeldAt, h.ExpiresAt,
	)
	metrics.ObserveQuery("CardHoldRepo.Create", start, err)
	return err
}

func (r *CardHoldRepository) GetByID(ctx context.Context, id string) (*card.Hold, error) {
	start := time.Now()
	h := &card.Hold{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, card_id, account_id, merchant, amount, currency, status, held_at, expires_at, captured_at, released_at
		 FROM card_holds WHERE id = $1`,
		id,
	).Scan(&h.ID, &h.CardID, &h.AccountID, &h.Merchant, &h.Amount, &h.Currency, &h.Status, &h.HeldAt, &h.ExpiresAt, &h.CapturedAt, &h.ReleasedAt)
	metrics.ObserveQuery("CardHoldRepo.GetByID", start, err)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (r *CardHoldRepository) ListByCardID(ctx context.Context, cardID string) ([]*card.Hold, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT id, card_id, account_id, merchant, amount, currency, status, held_at, expires_at, captured_at, released_at
		 FROM card_holds WHERE card_id = $1 ORDER BY held_at DESC`,
		cardID,
	)
	if err != nil {
		metrics.ObserveQuery("CardHoldRepo.ListByCardID", start, err)
		return nil, err
	}
	defer rows.Close()
	return r.scanHolds(rows, start)
}

func (r *CardHoldRepository) ListActiveByAccountID(ctx context.Context, accountID string) ([]*card.Hold, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT id, card_id, account_id, merchant, amount, currency, status, held_at, expires_at, captured_at, released_at
		 FROM card_holds WHERE account_id = $1 AND status = 'HELD' ORDER BY held_at DESC`,
		accountID,
	)
	if err != nil {
		metrics.ObserveQuery("CardHoldRepo.ListActiveByAccountID", start, err)
		return nil, err
	}
	defer rows.Close()
	return r.scanHolds(rows, start)
}

func (r *CardHoldRepository) Update(ctx context.Context, h *card.Hold) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE card_holds SET status = $1, captured_at = $2, released_at = $3 WHERE id = $4`,
		h.Status, h.CapturedAt, h.ReleasedAt, h.ID,
	)
	metrics.ObserveQuery("CardHoldRepo.Update", start, err)
	return err
}

func (r *CardHoldRepository) FetchExpired(ctx context.Context, limit int) ([]*card.Hold, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT id, card_id, account_id, merchant, amount, currency, status, held_at, expires_at, captured_at, released_at
		 FROM card_holds WHERE status = 'HELD' AND expires_at < NOW()
		 LIMIT $1 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		metrics.ObserveQuery("CardHoldRepo.FetchExpired", start, err)
		return nil, err
	}
	defer rows.Close()
	return r.scanHolds(rows, start)
}

func (r *CardHoldRepository) scanHolds(rows interface{ Next() bool; Scan(...interface{}) error }, start time.Time) ([]*card.Hold, error) {
	var holds []*card.Hold
	for rows.Next() {
		h := &card.Hold{}
		if err := rows.Scan(&h.ID, &h.CardID, &h.AccountID, &h.Merchant, &h.Amount, &h.Currency, &h.Status, &h.HeldAt, &h.ExpiresAt, &h.CapturedAt, &h.ReleasedAt); err != nil {
			return nil, err
		}
		holds = append(holds, h)
	}
	metrics.ObserveQuery("CardHoldRepo.scan", start, nil)
	return holds, nil
}
