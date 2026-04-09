package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/domain"
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

func (r *WriteRepo) Create(ctx context.Context, v *domain.Verification) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
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

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.Verification, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	v := &domain.Verification{}
	var reviewedBy *string
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE id = $1`, id,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &reviewedBy, &v.CreatedAt, &v.UpdatedAt)
	metrics.ObserveQuery("KYCRepo.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrKYCRequired
	}
	if reviewedBy != nil {
		v.ReviewedBy = *reviewedBy
	}
	return v, nil
}

func (r *WriteRepo) GetByUserID(ctx context.Context, userID string) (*domain.Verification, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	v := &domain.Verification{}
	var reviewedBy *string
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE user_id = $1`, userID,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &reviewedBy, &v.CreatedAt, &v.UpdatedAt)
	metrics.ObserveQuery("KYCRepo.GetByUserID", start, err)
	if err != nil {
		return nil, domain.ErrKYCRequired
	}
	if reviewedBy != nil {
		v.ReviewedBy = *reviewedBy
	}
	return v, nil
}

func (r *WriteRepo) Update(ctx context.Context, v *domain.Verification) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var reviewedBy interface{}
	if v.ReviewedBy != "" {
		reviewedBy = v.ReviewedBy
	}
	_, err := db.Exec(ctx,
		`UPDATE kyc_verifications SET status = $1, rejected_reason = $2, reviewed_by = $3, updated_at = $4 WHERE id = $5`,
		v.Status, v.RejectedReason, reviewedBy, v.UpdatedAt, v.ID,
	)
	metrics.ObserveQuery("KYCRepo.Update", start, err)
	return err
}

func (r *WriteRepo) ListPending(ctx context.Context, limit, offset int) ([]*domain.Verification, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE status = 'PENDING' ORDER BY created_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		metrics.ObserveQuery("KYCRepo.ListPending", start, err)
		return nil, fmt.Errorf("kyc_repo: list_pending: %w", err)
	}
	defer rows.Close()

	var items []*domain.Verification
	for rows.Next() {
		v := &domain.Verification{}
		var reviewedBy *string
		if err := rows.Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &reviewedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			metrics.ObserveQuery("KYCRepo.ListPending", start, err)
			return nil, fmt.Errorf("kyc_repo: list_pending scan: %w", err)
		}
		if reviewedBy != nil {
			v.ReviewedBy = *reviewedBy
		}
		items = append(items, v)
	}
	metrics.ObserveQuery("KYCRepo.ListPending", start, nil)
	return items, nil
}

func (r *WriteRepo) CountPending(ctx context.Context) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM kyc_verifications WHERE status = 'PENDING'`).Scan(&count)
	metrics.ObserveQuery("KYCRepo.CountPending", start, err)
	return count, err
}
