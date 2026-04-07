package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/beneficiary"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedPG "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
)

type BeneficiaryRepository struct {
	pool *pgxpool.Pool
}

func NewBeneficiaryRepository(pool *pgxpool.Pool) *BeneficiaryRepository {
	return &BeneficiaryRepository{pool: pool}
}

func (r *BeneficiaryRepository) Create(ctx context.Context, b *beneficiary.Beneficiary) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO beneficiaries (user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		b.UserID, b.Name, b.AccountNumber, b.BankName, b.BankCode, b.Currency, b.BenType, b.Verified, b.CreatedAt,
	).Scan(&b.ID)
	metrics.ObserveQuery("BeneficiaryRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("beneficiary_repo: create: %w", err)
	}
	return nil
}

func (r *BeneficiaryRepository) GetByID(ctx context.Context, id string) (*beneficiary.Beneficiary, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	b := &beneficiary.Beneficiary{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at
		 FROM beneficiaries WHERE id = $1`, id,
	).Scan(&b.ID, &b.UserID, &b.Name, &b.AccountNumber, &b.BankName, &b.BankCode, &b.Currency, &b.BenType, &b.Verified, &b.CreatedAt)
	metrics.ObserveQuery("BeneficiaryRepo.GetByID", start, err)
	if err != nil {
		return nil, beneficiary.ErrBeneficiaryNotFound
	}
	return b, nil
}

func (r *BeneficiaryRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*beneficiary.Beneficiary, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, name, account_number, bank_name, bank_code, currency, ben_type, verified, created_at
		 FROM beneficiaries WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("BeneficiaryRepo.ListByUserID", start, err)
		return nil, fmt.Errorf("beneficiary_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*beneficiary.Beneficiary
	for rows.Next() {
		b := &beneficiary.Beneficiary{}
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.AccountNumber, &b.BankName, &b.BankCode, &b.Currency, &b.BenType, &b.Verified, &b.CreatedAt); err != nil {
			metrics.ObserveQuery("BeneficiaryRepo.ListByUserID", start, err)
			return nil, fmt.Errorf("beneficiary_repo: list scan: %w", err)
		}
		items = append(items, b)
	}
	metrics.ObserveQuery("BeneficiaryRepo.ListByUserID", start, nil)
	return items, nil
}

func (r *BeneficiaryRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM beneficiaries WHERE user_id = $1`, userID).Scan(&count)
	metrics.ObserveQuery("BeneficiaryRepo.CountByUserID", start, err)
	return count, err
}

func (r *BeneficiaryRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM beneficiaries WHERE id = $1`, id)
	metrics.ObserveQuery("BeneficiaryRepo.Delete", start, err)
	return err
}

func (r *BeneficiaryRepository) ExistsByUserAndAccount(ctx context.Context, userID, accountNumber string) (bool, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM beneficiaries WHERE user_id = $1 AND account_number = $2)`,
		userID, accountNumber,
	).Scan(&exists)
	metrics.ObserveQuery("BeneficiaryRepo.ExistsByUserAndAccount", start, err)
	return exists, err
}
