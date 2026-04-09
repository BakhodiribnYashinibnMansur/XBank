package account_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
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
	user, err := userdomain.NewUser("test@bank.com", "$2a$12$hash", "Test", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestAccountRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	acc, err := domain.NewAccount(userID, kerneldomain.UZS)
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	acc.ClearUncommittedEvents()

	if err := repo.Create(ctx, acc); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if acc.ID == "" {
		t.Fatal("expected account ID to be set")
	}

	got, err := repo.GetByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}

	if got.UserID != userID {
		t.Errorf("userID = %q, want %q", got.UserID, userID)
	}
	if got.Balance.Amount != 0 {
		t.Errorf("balance = %d, want 0", got.Balance.Amount)
	}
	if got.Balance.Currency != kerneldomain.UZS {
		t.Errorf("currency = %q, want %q", got.Balance.Currency, kerneldomain.UZS)
	}
	if got.Status != domain.StatusActive {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusActive)
	}
}

func TestAccountRepo_ListByUserID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	for i := 0; i < 3; i++ {
		acc, _ := domain.NewAccount(userID, kerneldomain.UZS)
		acc.ClearUncommittedEvents()
		if err := repo.Create(ctx, acc); err != nil {
			t.Fatalf("creating account %d: %v", i, err)
		}
	}

	accounts, err := repo.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(accounts) != 3 {
		t.Errorf("got %d accounts, want 3", len(accounts))
	}
}

func TestAccountRepo_CountByUserID(t *testing.T) {
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

	acc, _ := domain.NewAccount(userID, kerneldomain.UZS)
	acc.ClearUncommittedEvents()
	repo.Create(ctx, acc)

	count, _ = repo.CountByUserID(ctx, userID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestAccountRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	acc, _ := domain.NewAccount(userID, kerneldomain.UZS)
	acc.ClearUncommittedEvents()
	repo.Create(ctx, acc)

	acc.Balance.Amount = 100_000
	acc.Status = domain.StatusFrozen
	if err := repo.Update(ctx, acc); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, acc.ID)
	if got.Balance.Amount != 100_000 {
		t.Errorf("balance = %d, want 100000", got.Balance.Amount)
	}
	if got.Status != domain.StatusFrozen {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusFrozen)
	}
}

func TestAccountRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent account")
	}
}
