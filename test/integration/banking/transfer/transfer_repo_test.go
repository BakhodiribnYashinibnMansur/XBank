package transfer_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	accdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	accpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/infrastructure/postgres"
	userdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	userpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
	kerneldomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
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
	user, err := userdomain.NewUser("transfer@bank.com", "$2a$12$hash", "Transfer", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func createTestAccount(t *testing.T, userID string) string {
	t.Helper()
	accRepo := accpg.NewWriteRepo(pgc.Pool)
	acc, err := accdomain.NewAccount(userID, kerneldomain.UZS)
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	acc.ClearUncommittedEvents()
	if err := accRepo.Create(context.Background(), acc); err != nil {
		t.Fatalf("inserting test account: %v", err)
	}
	return acc.ID
}

func TestTransferRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	amount := kerneldomain.Money{Amount: 50000, Currency: kerneldomain.UZS}
	tr, err := domain.NewTransfer(fromAccID, toAccID, amount, "test transfer")
	if err != nil {
		t.Fatalf("NewTransfer: %v", err)
	}
	tr.ClearUncommittedEvents()

	if err := repo.Create(ctx, tr); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	got, err := repo.GetByID(ctx, tr.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.FromAccountID != fromAccID {
		t.Errorf("FromAccountID = %q, want %q", got.FromAccountID, fromAccID)
	}
	if got.ToAccountID != toAccID {
		t.Errorf("ToAccountID = %q, want %q", got.ToAccountID, toAccID)
	}
	if got.Amount.Amount != 50000 {
		t.Errorf("Amount = %d, want 50000", got.Amount.Amount)
	}
	if got.Status != domain.StatusPending {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusPending)
	}
	if got.Description != "test transfer" {
		t.Errorf("Description = %q, want %q", got.Description, "test transfer")
	}
}

func TestTransferRepo_ListByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	for i := 0; i < 3; i++ {
		amount := kerneldomain.Money{Amount: int64(1000 * (i + 1)), Currency: kerneldomain.UZS}
		tr, _ := domain.NewTransfer(fromAccID, toAccID, amount, fmt.Sprintf("transfer %d", i))
		tr.ClearUncommittedEvents()
		if err := repo.Create(ctx, tr); err != nil {
			t.Fatalf("creating transfer %d: %v", i, err)
		}
	}

	// From account perspective
	transfers, err := repo.ListByAccountID(ctx, fromAccID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAccountID: %v", err)
	}
	if len(transfers) != 3 {
		t.Errorf("got %d transfers, want 3", len(transfers))
	}

	// To account perspective should also see them
	transfers, err = repo.ListByAccountID(ctx, toAccID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAccountID (to): %v", err)
	}
	if len(transfers) != 3 {
		t.Errorf("got %d transfers from to account, want 3", len(transfers))
	}
}

func TestTransferRepo_CountByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	count, err := repo.CountByAccountID(ctx, fromAccID)
	if err != nil {
		t.Fatalf("CountByAccountID: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	amount := kerneldomain.Money{Amount: 10000, Currency: kerneldomain.UZS}
	tr, _ := domain.NewTransfer(fromAccID, toAccID, amount, "count test")
	tr.ClearUncommittedEvents()
	repo.Create(ctx, tr)

	count, _ = repo.CountByAccountID(ctx, fromAccID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestTransferRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	amount := kerneldomain.Money{Amount: 25000, Currency: kerneldomain.UZS}
	tr, _ := domain.NewTransfer(fromAccID, toAccID, amount, "update test")
	tr.ClearUncommittedEvents()
	repo.Create(ctx, tr)

	tr.Status = domain.StatusFailed
	tr.FailureReason = "insufficient funds"
	if err := repo.Update(ctx, tr); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, tr.ID)
	if got.Status != domain.StatusFailed {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusFailed)
	}
	if got.FailureReason != "insufficient funds" {
		t.Errorf("FailureReason = %q, want %q", got.FailureReason, "insufficient funds")
	}
}

func TestTransferRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent transfer")
	}
}
