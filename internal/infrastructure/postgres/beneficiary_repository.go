package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/beneficiary"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BeneficiaryRepository struct {
	pool *pgxpool.Pool
}

func NewBeneficiaryRepository(pool *pgxpool.Pool) *BeneficiaryRepository {
	return &BeneficiaryRepository{pool: pool}
}

func (r *BeneficiaryRepository) Create(ctx context.Context, b *beneficiary.Beneficiary) error {
	db := ExtractDBTX(ctx, r.pool)
	return db.QueryRow(ctx,
		`INSERT INTO beneficiaries (user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		b.UserID, b.Name, b.AccountNumber, b.BankName, b.BankCode, b.Currency, b.BenType, b.Verified, b.CreatedAt,
	).Scan(&b.ID)
}

func (r *BeneficiaryRepository) GetByID(ctx context.Context, id string) (*beneficiary.Beneficiary, error) {
	db := ExtractDBTX(ctx, r.pool)
	b := &beneficiary.Beneficiary{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at
		 FROM beneficiaries WHERE id = $1`, id,
	).Scan(&b.ID, &b.UserID, &b.Name, &b.AccountNumber, &b.BankName, &b.BankCode, &b.Currency, &b.BenType, &b.Verified, &b.CreatedAt)
	if err != nil {
		return nil, beneficiary.ErrBeneficiaryNotFound
	}
	return b, nil
}

func (r *BeneficiaryRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*beneficiary.Beneficiary, error) {
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at
		 FROM beneficiaries WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*beneficiary.Beneficiary
	for rows.Next() {
		b := &beneficiary.Beneficiary{}
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.AccountNumber, &b.BankName, &b.BankCode, &b.Currency, &b.BenType, &b.Verified, &b.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return items, nil
}

func (r *BeneficiaryRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM beneficiaries WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *BeneficiaryRepository) Delete(ctx context.Context, id string) error {
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM beneficiaries WHERE id = $1`, id)
	return err
}

func (r *BeneficiaryRepository) ExistsByUserAndAccount(ctx context.Context, userID, accountNumber string) (bool, error) {
	db := ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM beneficiaries WHERE user_id = $1 AND account_number = $2)`,
		userID, accountNumber,
	).Scan(&exists)
	return exists, err
}
