package ledger_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	accdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	accpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/infrastructure/postgres"
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
	user, err := userdomain.NewUser("ledger@bank.com", "$2a$12$hash", "Ledger", "User")
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

func TestLedgerRepo_CreatePairAndList(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	transferID := "a0000000-0000-4000-8000-000000000001"
	debit, credit := domain.CreatePair(transferID, fromAccID, toAccID, 50000, "UZS")

	if err := repo.CreatePair(ctx, debit, credit); err != nil {
		t.Fatalf("CreatePair: %v", err)
	}
	if debit.ID == "" {
		t.Fatal("expected debit ID to be set")
	}
	if credit.ID == "" {
		t.Fatal("expected credit ID to be set")
	}

	// List from sender's perspective
	entries, err := repo.ListByAccountID(ctx, fromAccID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAccountID: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d entries for sender, want 1", len(entries))
	}
	if entries[0].EntryType != domain.Debit {
		t.Errorf("entry type = %q, want %q", entries[0].EntryType, domain.Debit)
	}

	// List from receiver's perspective
	entries, err = repo.ListByAccountID(ctx, toAccID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAccountID receiver: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d entries for receiver, want 1", len(entries))
	}
	if entries[0].EntryType != domain.Credit {
		t.Errorf("entry type = %q, want %q", entries[0].EntryType, domain.Credit)
	}
}

func TestLedgerRepo_CountByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accID := createTestAccount(t, userID)

	count, err := repo.CountByAccountID(ctx, accID)
	if err != nil {
		t.Fatalf("CountByAccountID: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	toAccID := createTestAccount(t, userID)
	debit, credit := domain.CreatePair("b0000000-0000-4000-8000-000000000001", accID, toAccID, 10000, "UZS")
	repo.CreatePair(ctx, debit, credit)

	count, _ = repo.CountByAccountID(ctx, accID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestLedgerRepo_BalanceByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accA := createTestAccount(t, userID)
	accB := createTestAccount(t, userID)

	// Transfer 100 from A to B
	d1, c1 := domain.CreatePair("c0000000-0000-4000-8000-000000000001", accA, accB, 10000, "UZS")
	repo.CreatePair(ctx, d1, c1)

	// Transfer 30 from B to A
	d2, c2 := domain.CreatePair("c0000000-0000-4000-8000-000000000002", accB, accA, 3000, "UZS")
	repo.CreatePair(ctx, d2, c2)

	// A balance: -10000 (debit) + 3000 (credit) = -7000
	balanceA, err := repo.BalanceByAccountID(ctx, accA)
	if err != nil {
		t.Fatalf("BalanceByAccountID A: %v", err)
	}
	if balanceA != -7000 {
		t.Errorf("balance A = %d, want -7000", balanceA)
	}

	// B balance: +10000 (credit) - 3000 (debit) = +7000
	balanceB, err := repo.BalanceByAccountID(ctx, accB)
	if err != nil {
		t.Fatalf("BalanceByAccountID B: %v", err)
	}
	if balanceB != 7000 {
		t.Errorf("balance B = %d, want 7000", balanceB)
	}
}

func TestLedgerRepo_BalanceByAccountID_Empty(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accID := createTestAccount(t, userID)

	balance, err := repo.BalanceByAccountID(ctx, accID)
	if err != nil {
		t.Fatalf("BalanceByAccountID: %v", err)
	}
	if balance != 0 {
		t.Errorf("balance = %d, want 0 for empty account", balance)
	}
}
