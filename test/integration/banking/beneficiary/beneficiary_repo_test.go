package beneficiary_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/infrastructure/postgres"
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

func createTestUser(t *testing.T) string {
	t.Helper()
	userRepo := userpg.NewWriteRepo(pgc.Pool)
	user, err := userdomain.NewUser("ben@bank.com", "$2a$12$hash", "Ben", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestBeneficiaryRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	ben, err := domain.NewBeneficiary(userID, "John Doe", "UZ1234567890123456", "National Bank", "NBUUUZ2X", "UZS", domain.TypeInternal)
	if err != nil {
		t.Fatalf("NewBeneficiary: %v", err)
	}

	if err := repo.Create(ctx, ben); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if ben.ID == "" {
		t.Fatal("expected beneficiary ID to be set")
	}

	got, err := repo.GetByID(ctx, ben.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.Name != "John Doe" {
		t.Errorf("Name = %q, want %q", got.Name, "John Doe")
	}
	if got.AccountNumber != "UZ1234567890123456" {
		t.Errorf("AccountNumber = %q, want %q", got.AccountNumber, "UZ1234567890123456")
	}
	if got.BenType != domain.TypeInternal {
		t.Errorf("BenType = %q, want %q", got.BenType, domain.TypeInternal)
	}
	if got.Verified != false {
		t.Errorf("Verified = %v, want false", got.Verified)
	}
}

func TestBeneficiaryRepo_ListByUserID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	for i := 0; i < 3; i++ {
		ben, _ := domain.NewBeneficiary(userID, fmt.Sprintf("Recipient %d", i), fmt.Sprintf("ACC%d", i), "Bank", "", "UZS", domain.TypeInternal)
		if err := repo.Create(ctx, ben); err != nil {
			t.Fatalf("creating beneficiary %d: %v", i, err)
		}
	}

	items, err := repo.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("got %d beneficiaries, want 3", len(items))
	}
}

func TestBeneficiaryRepo_CountByUserID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	count, err := repo.CountByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("CountByUserID: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	ben, _ := domain.NewBeneficiary(userID, "Test", "ACC001", "Bank", "", "UZS", domain.TypeInternal)
	repo.Create(ctx, ben)

	count, _ = repo.CountByUserID(ctx, userID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestBeneficiaryRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	ben, _ := domain.NewBeneficiary(userID, "To Delete", "ACC-DEL", "Bank", "", "UZS", domain.TypeInternal)
	repo.Create(ctx, ben)

	if err := repo.Delete(ctx, ben.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, ben.ID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestBeneficiaryRepo_ExistsByUserAndAccount(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	exists, err := repo.ExistsByUserAndAccount(ctx, userID, "ACC-EXISTS")
	if err != nil {
		t.Fatalf("ExistsByUserAndAccount: %v", err)
	}
	if exists {
		t.Error("expected false before creation")
	}

	ben, _ := domain.NewBeneficiary(userID, "Exists Check", "ACC-EXISTS", "Bank", "", "UZS", domain.TypeInternal)
	repo.Create(ctx, ben)

	exists, err = repo.ExistsByUserAndAccount(ctx, userID, "ACC-EXISTS")
	if err != nil {
		t.Fatalf("ExistsByUserAndAccount after create: %v", err)
	}
	if !exists {
		t.Error("expected true after creation")
	}
}

func TestBeneficiaryRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent beneficiary")
	}
}
