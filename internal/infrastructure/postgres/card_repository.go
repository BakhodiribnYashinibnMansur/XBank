package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	pool *pgxpool.Pool
}

func NewCardRepository(pool *pgxpool.Pool) *CardRepository {
	return &CardRepository{pool: pool}
}

func (r *CardRepository) Create(ctx context.Context, c *card.Card) error {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO cards (account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		                   card_type, status, pin_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	return db.QueryRow(ctx, query,
		c.AccountID, c.PAN, c.MaskedPAN, c.PINHash, c.ExpiryMonth, c.ExpiryYear,
		c.CardType, c.Status, c.PINAttempts, c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*card.Card, error) {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		       card_type, status, pin_attempts, created_at, updated_at
		FROM cards WHERE id = $1`

	c := &card.Card{}
	err := db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.AccountID, &c.PAN, &c.MaskedPAN, &c.PINHash,
		&c.ExpiryMonth, &c.ExpiryYear, &c.CardType, &c.Status,
		&c.PINAttempts, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, card.ErrCardNotFound
	}
	return c, nil
}

func (r *CardRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*card.Card, error) {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		       card_type, status, pin_attempts, created_at, updated_at
		FROM cards WHERE account_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		c := &card.Card{}
		if err := rows.Scan(
			&c.ID, &c.AccountID, &c.PAN, &c.MaskedPAN, &c.PINHash,
			&c.ExpiryMonth, &c.ExpiryYear, &c.CardType, &c.Status,
			&c.PINAttempts, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM cards WHERE account_id = $1`, accountID).Scan(&count)
	return count, err
}

func (r *CardRepository) Update(ctx context.Context, c *card.Card) error {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		UPDATE cards SET pin_hash = $1, status = $2, pin_attempts = $3, updated_at = $4
		WHERE id = $5`

	_, err := db.Exec(ctx, query, c.PINHash, c.Status, c.PINAttempts, c.UpdatedAt, c.ID)
	return err
}
