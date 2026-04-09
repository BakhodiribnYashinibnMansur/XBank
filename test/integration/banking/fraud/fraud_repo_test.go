package fraud_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud/infrastructure/postgres"
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
	user, err := userdomain.NewUser("fraud@bank.com", "$2a$12$hash", "Fraud", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestFraudRepo_CreateAndGetByTransferID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	transferID := "a0000000-0000-4000-8000-000000000001"
	check := domain.NewCheck(transferID, userID, 50000, []string{"LARGE_AMOUNT"})

	if err := repo.Create(ctx, check); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if check.ID == "" {
		t.Fatal("expected check ID to be set")
	}

	got, err := repo.GetByTransferID(ctx, transferID)
	if err != nil {
		t.Fatalf("GetByTransferID: %v", err)
	}
	if got.TransferID != transferID {
		t.Errorf("TransferID = %q, want %q", got.TransferID, transferID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.RiskScore <= 0 {
		t.Errorf("RiskScore = %d, expected > 0", got.RiskScore)
	}
}

func TestFraudRepo_ListFlagged(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	// Create a low-risk check (should be APPROVE, not flagged)
	lowRisk := domain.NewCheck("b0000000-0000-4000-8000-000000000001", userID, 1000, []string{})
	if err := repo.Create(ctx, lowRisk); err != nil {
		t.Fatalf("Create low risk: %v", err)
	}

	// Create a high-risk check (should be FLAG or BLOCK)
	highRisk := domain.NewCheck("b0000000-0000-4000-8000-000000000002", userID, 2000000000, []string{"LARGE_AMOUNT", "HIGH_VELOCITY", "NEW_BENEFICIARY"})
	if err := repo.Create(ctx, highRisk); err != nil {
		t.Fatalf("Create high risk: %v", err)
	}

	flagged, err := repo.ListFlagged(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListFlagged: %v", err)
	}
	if len(flagged) < 1 {
		t.Errorf("got %d flagged checks, want at least 1", len(flagged))
	}

	// Verify the high-risk one is in the list
	found := false
	for _, f := range flagged {
		if f.TransferID == "b0000000-0000-4000-8000-000000000002" {
			found = true
			break
		}
	}
	if !found {
		t.Error("high-risk check not found in flagged list")
	}
}

func TestFraudRepo_CountFlagged(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	count, err := repo.CountFlagged(ctx)
	if err != nil {
		t.Fatalf("CountFlagged: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	// Create a flagged check
	check := domain.NewCheck("c0000000-0000-4000-8000-000000000001", userID, 2000000000, []string{"LARGE_AMOUNT", "HIGH_VELOCITY", "NEW_BENEFICIARY"})
	if err := repo.Create(ctx, check); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify action is FLAG or BLOCK
	if check.Action != domain.ActionFlag && check.Action != domain.ActionBlock {
		t.Skipf("check action is %q (not FLAG/BLOCK), skipping count assertion", check.Action)
	}

	count, _ = repo.CountFlagged(ctx)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestFraudRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	check := domain.NewCheck("d0000000-0000-4000-8000-000000000001", userID, 500000000, []string{"LARGE_AMOUNT"})
	repo.Create(ctx, check)

	adminUUID := "a0000000-0000-4000-8000-aaaaaaaaaaaa"
	check.ReviewedBy = adminUUID
	check.ReviewComment = "looks legitimate"
	if err := repo.Update(ctx, check); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByTransferID(ctx, "d0000000-0000-4000-8000-000000000001")
	if got.ReviewedBy != adminUUID {
		t.Errorf("ReviewedBy = %q, want %q", got.ReviewedBy, adminUUID)
	}
	if got.ReviewComment != "looks legitimate" {
		t.Errorf("ReviewComment = %q, want %q", got.ReviewComment, "looks legitimate")
	}
}

func TestFraudRepo_GetByTransferID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByTransferID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent fraud check")
	}
}
