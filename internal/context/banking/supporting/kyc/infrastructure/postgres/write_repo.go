package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/kyc"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KYCRepository struct {
	pool *pgxpool.Pool
}

func NewKYCRepository(pool *pgxpool.Pool) *KYCRepository {
	return &KYCRepository{pool: pool}
}

func (r *KYCRepository) Create(ctx context.Context, v *kyc.Verification) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO kyc_verifications (user_id, document_type, document_number, first_name, last_name, date_of_birth, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		v.UserID, v.DocumentType, v.DocumentNumber, v.FirstName, v.LastName, v.DateOfBirth, v.Status, v.CreatedAt, v.UpdatedAt,
	).Scan(&v.ID)
	metrics.ObserveQuery("KYCRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("kyc_repo: create: %w", err)
	}
	return nil
}

func (r *KYCRepository) GetByID(ctx context.Context, id string) (*kyc.Verification, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	v := &kyc.Verification{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE id = $1`, id,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt)
	metrics.ObserveQuery("KYCRepo.GetByID", start, err)
	if err != nil {
		return nil, kyc.ErrKYCRequired
	}
	return v, nil
}

func (r *KYCRepository) GetByUserID(ctx context.Context, userID string) (*kyc.Verification, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	v := &kyc.Verification{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE user_id = $1`, userID,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt)
	metrics.ObserveQuery("KYCRepo.GetByUserID", start, err)
	if err != nil {
		return nil, kyc.ErrKYCRequired
	}
	return v, nil
}

func (r *KYCRepository) Update(ctx context.Context, v *kyc.Verification) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE kyc_verifications SET status = $1, rejected_reason = $2, reviewed_by = $3, updated_at = $4 WHERE id = $5`,
		v.Status, v.RejectedReason, v.ReviewedBy, v.UpdatedAt, v.ID,
	)
	metrics.ObserveQuery("KYCRepo.Update", start, err)
	return err
}

func (r *KYCRepository) ListPending(ctx context.Context, limit, offset int) ([]*kyc.Verification, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE status = 'PENDING' ORDER BY created_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		metrics.ObserveQuery("KYCRepo.ListPending", start, err)
		return nil, fmt.Errorf("kyc_repo: list_pending: %w", err)
	}
	defer rows.Close()

	var items []*kyc.Verification
	for rows.Next() {
		v := &kyc.Verification{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			metrics.ObserveQuery("KYCRepo.ListPending", start, err)
			return nil, fmt.Errorf("kyc_repo: list_pending scan: %w", err)
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("KYCRepo.ListPending", start, nil)
	return items, nil
}

func (r *KYCRepository) CountPending(ctx context.Context) (int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM kyc_verifications WHERE status = 'PENDING'`).Scan(&count)
	metrics.ObserveQuery("KYCRepo.CountPending", start, err)
	return count, err
}
