package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/kyc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KYCRepository struct {
	pool *pgxpool.Pool
}

func NewKYCRepository(pool *pgxpool.Pool) *KYCRepository {
	return &KYCRepository{pool: pool}
}

func (r *KYCRepository) Create(ctx context.Context, v *kyc.Verification) error {
	db := ExtractDBTX(ctx, r.pool)
	return db.QueryRow(ctx,
		`INSERT INTO kyc_verifications (user_id, document_type, document_number, first_name, last_name, date_of_birth, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		v.UserID, v.DocumentType, v.DocumentNumber, v.FirstName, v.LastName, v.DateOfBirth, v.Status, v.CreatedAt, v.UpdatedAt,
	).Scan(&v.ID)
}

func (r *KYCRepository) GetByID(ctx context.Context, id string) (*kyc.Verification, error) {
	db := ExtractDBTX(ctx, r.pool)
	v := &kyc.Verification{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE id = $1`, id,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, kyc.ErrKYCRequired
	}
	return v, nil
}

func (r *KYCRepository) GetByUserID(ctx context.Context, userID string) (*kyc.Verification, error) {
	db := ExtractDBTX(ctx, r.pool)
	v := &kyc.Verification{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE user_id = $1`, userID,
	).Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, kyc.ErrKYCRequired
	}
	return v, nil
}

func (r *KYCRepository) Update(ctx context.Context, v *kyc.Verification) error {
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE kyc_verifications SET status = $1, rejected_reason = $2, reviewed_by = $3, updated_at = $4 WHERE id = $5`,
		v.Status, v.RejectedReason, v.ReviewedBy, v.UpdatedAt, v.ID,
	)
	return err
}

func (r *KYCRepository) ListPending(ctx context.Context, limit, offset int) ([]*kyc.Verification, error) {
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at
		 FROM kyc_verifications WHERE status = 'PENDING' ORDER BY created_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*kyc.Verification
	for rows.Next() {
		v := &kyc.Verification{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.DocumentType, &v.DocumentNumber, &v.FirstName, &v.LastName, &v.DateOfBirth, &v.Status, &v.RejectedReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, nil
}

func (r *KYCRepository) CountPending(ctx context.Context) (int64, error) {
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM kyc_verifications WHERE status = 'PENDING'`).Scan(&count)
	return count, err
}
