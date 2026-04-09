package kyc_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/infrastructure/postgres"
	userdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	userpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
)

var pgc *setup.PostgresContainer

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping integration tests (set INTEGRATION=1 to run)")
		os.Exit(0)
	}

	var code int
	func() {
		t := &testing.T{}
		pgc = setup.MustPostgres(t)
		defer pgc.Teardown()
		code = m.Run()
	}()
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	setup.TruncateAll(t, pgc.Pool)
}

func createTestUser(t *testing.T, email string) string {
	t.Helper()
	userRepo := userpg.NewWriteRepo(pgc.Pool)
	user, err := userdomain.NewUser(email, "$2a$12$hash", "KYC", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestKYCRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t, "kyc1@bank.com")

	v, err := domain.NewVerification(userID, domain.DocPassport, "AB1234567", "John", "Doe", "1990-01-15")
	if err != nil {
		t.Fatalf("NewVerification: %v", err)
	}

	if err := repo.Create(ctx, v); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if v.ID == "" {
		t.Fatal("expected verification ID to be set")
	}

	got, err := repo.GetByID(ctx, v.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.DocumentType != domain.DocPassport {
		t.Errorf("DocumentType = %q, want %q", got.DocumentType, domain.DocPassport)
	}
	if got.DocumentNumber != "AB1234567" {
		t.Errorf("DocumentNumber = %q, want %q", got.DocumentNumber, "AB1234567")
	}
	if got.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", got.FirstName, "John")
	}
	if got.Status != domain.StatusPending {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusPending)
	}
}

func TestKYCRepo_GetByUserID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t, "kyc2@bank.com")

	v, _ := domain.NewVerification(userID, domain.DocIDCard, "ID9876543", "Jane", "Smith", "1985-06-20")
	repo.Create(ctx, v)

	got, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if got.ID != v.ID {
		t.Errorf("ID = %q, want %q", got.ID, v.ID)
	}
}

func TestKYCRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t, "kyc3@bank.com")

	v, _ := domain.NewVerification(userID, domain.DocPassport, "PP1111111", "Ali", "Veli", "1995-03-10")
	repo.Create(ctx, v)

	adminUUID := "a0000000-0000-4000-8000-000000000001"
	v.Approve(adminUUID)
	if err := repo.Update(ctx, v); err != nil {
		t.Fatalf("Update (approve): %v", err)
	}

	got, _ := repo.GetByID(ctx, v.ID)
	if got.Status != domain.StatusApproved {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusApproved)
	}
	if got.ReviewedBy != adminUUID {
		t.Errorf("ReviewedBy = %q, want %q", got.ReviewedBy, adminUUID)
	}
}

func TestKYCRepo_Update_Reject(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t, "kyc4@bank.com")

	v, _ := domain.NewVerification(userID, domain.DocDriverLicense, "DL5555555", "Bob", "Test", "2000-12-25")
	repo.Create(ctx, v)

	v.Reject("a0000000-0000-4000-8000-000000000002", "document expired")
	if err := repo.Update(ctx, v); err != nil {
		t.Fatalf("Update (reject): %v", err)
	}

	got, _ := repo.GetByID(ctx, v.ID)
	if got.Status != domain.StatusRejected {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusRejected)
	}
	if got.RejectedReason != "document expired" {
		t.Errorf("RejectedReason = %q, want %q", got.RejectedReason, "document expired")
	}
}

func TestKYCRepo_ListPending(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	// Create 3 pending verifications
	for i := 0; i < 3; i++ {
		userID := createTestUser(t, fmt.Sprintf("kyc-list-%d@bank.com", i))
		v, _ := domain.NewVerification(userID, domain.DocPassport, fmt.Sprintf("PP%07d", i), "User", fmt.Sprintf("N%d", i), "1990-01-01")
		repo.Create(ctx, v)
	}

	// Approve one
	userID4 := createTestUser(t, "kyc-list-approved@bank.com")
	v4, _ := domain.NewVerification(userID4, domain.DocPassport, "PP9999999", "Approved", "User", "1990-01-01")
	repo.Create(ctx, v4)
	v4.Approve("a0000000-0000-4000-8000-000000000099")
	repo.Update(ctx, v4)

	pending, err := repo.ListPending(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(pending) != 3 {
		t.Errorf("got %d pending, want 3", len(pending))
	}
}

func TestKYCRepo_CountPending(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	count, err := repo.CountPending(ctx)
	if err != nil {
		t.Fatalf("CountPending: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	userID := createTestUser(t, "kyc-count@bank.com")
	v, _ := domain.NewVerification(userID, domain.DocPassport, "PP0000001", "Count", "Test", "1990-01-01")
	repo.Create(ctx, v)

	count, _ = repo.CountPending(ctx)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestKYCRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent KYC verification")
	}
}

func TestKYCRepo_GetByUserID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent user KYC")
	}
}
