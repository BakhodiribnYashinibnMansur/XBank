package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CardRepository struct {
	pool *pgxpool.Pool
}

func NewCardRepository(pool *pgxpool.Pool) *CardRepository {
	return &CardRepository{pool: pool}
}

func (r *CardRepository) Create(ctx context.Context, c *card.Card) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO cards (account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		                   card_type, status, pin_attempts, three_ds_enrolled, three_ds_version, emv_aid,
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`

	err := db.QueryRow(ctx, query,
		c.AccountID, c.PAN, c.MaskedPAN, c.PINHash, c.ExpiryMonth, c.ExpiryYear,
		c.CardType, c.Status, c.PINAttempts, c.ThreeDSEnrolled, c.ThreeDSVersion, c.EMVAID,
		c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
	metrics.ObserveQuery("CardRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("card_repo: create: %w", err)
	}
	return nil
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*card.Card, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		       card_type, status, pin_attempts, three_ds_enrolled, three_ds_version, emv_aid,
		       created_at, updated_at
		FROM cards WHERE id = $1`

	c := &card.Card{}
	err := db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.AccountID, &c.PAN, &c.MaskedPAN, &c.PINHash,
		&c.ExpiryMonth, &c.ExpiryYear, &c.CardType, &c.Status,
		&c.PINAttempts, &c.ThreeDSEnrolled, &c.ThreeDSVersion, &c.EMVAID,
		&c.CreatedAt, &c.UpdatedAt,
	)
	metrics.ObserveQuery("CardRepo.GetByID", start, err)
	if err != nil {
		return nil, card.ErrCardNotFound
	}
	return c, nil
}

func (r *CardRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*card.Card, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year,
		       card_type, status, pin_attempts, three_ds_enrolled, three_ds_version, emv_aid,
		       created_at, updated_at
		FROM cards WHERE account_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		metrics.ObserveQuery("CardRepo.ListByAccountID", start, err)
		return nil, fmt.Errorf("card_repo: list: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		c := &card.Card{}
		if err := rows.Scan(
			&c.ID, &c.AccountID, &c.PAN, &c.MaskedPAN, &c.PINHash,
			&c.ExpiryMonth, &c.ExpiryYear, &c.CardType, &c.Status,
			&c.PINAttempts, &c.ThreeDSEnrolled, &c.ThreeDSVersion, &c.EMVAID,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			metrics.ObserveQuery("CardRepo.ListByAccountID", start, err)
			return nil, fmt.Errorf("card_repo: list scan: %w", err)
		}
		cards = append(cards, c)
	}
	metrics.ObserveQuery("CardRepo.ListByAccountID", start, nil)
	return cards, nil
}

func (r *CardRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM cards WHERE account_id = $1`, accountID).Scan(&count)
	metrics.ObserveQuery("CardRepo.CountByAccountID", start, err)
	return count, err
}

func (r *CardRepository) Update(ctx context.Context, c *card.Card) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		UPDATE cards SET pin_hash = $1, status = $2, pin_attempts = $3,
		       three_ds_enrolled = $4, three_ds_version = $5, emv_aid = $6, updated_at = $7
		WHERE id = $8`

	_, err := db.Exec(ctx, query, c.PINHash, c.Status, c.PINAttempts,
		c.ThreeDSEnrolled, c.ThreeDSVersion, c.EMVAID, c.UpdatedAt, c.ID)
	metrics.ObserveQuery("CardRepo.Update", start, err)
	return err
}
