package card_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	accdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	accpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/infrastructure/postgres"
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
	user, err := userdomain.NewUser("cardtest@bank.com", "$2a$12$hash", "Card", "User")
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

func TestCardRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accountID := createTestAccount(t, userID)

	card, err := domain.NewCard(accountID, domain.TypeDebit)
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	if err := repo.Create(ctx, card); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if card.ID == "" {
		t.Fatal("expected card ID to be set")
	}

	got, err := repo.GetByID(ctx, card.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.AccountID != accountID {
		t.Errorf("AccountID = %q, want %q", got.AccountID, accountID)
	}
	if got.CardType != domain.TypeDebit {
		t.Errorf("CardType = %q, want %q", got.CardType, domain.TypeDebit)
	}
	if got.Status != domain.StatusInactive {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusInactive)
	}
	if got.MaskedPAN == "" {
		t.Error("expected MaskedPAN to be set")
	}
}

func TestCardRepo_ListByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accountID := createTestAccount(t, userID)

	for i := 0; i < 3; i++ {
		card, _ := domain.NewCard(accountID, domain.TypeDebit)
		if err := repo.Create(ctx, card); err != nil {
			t.Fatalf("creating card %d: %v", i, err)
		}
	}

	cards, err := repo.ListByAccountID(ctx, accountID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAccountID: %v", err)
	}
	if len(cards) != 3 {
		t.Errorf("got %d cards, want 3", len(cards))
	}
}

func TestCardRepo_CountByAccountID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accountID := createTestAccount(t, userID)

	count, err := repo.CountByAccountID(ctx, accountID)
	if err != nil {
		t.Fatalf("CountByAccountID: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	card, _ := domain.NewCard(accountID, domain.TypeVirtual)
	repo.Create(ctx, card)

	count, _ = repo.CountByAccountID(ctx, accountID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestCardRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	accountID := createTestAccount(t, userID)

	card, _ := domain.NewCard(accountID, domain.TypeDebit)
	repo.Create(ctx, card)

	card.Status = domain.StatusBlocked
	card.PINAttempts = 3
	if err := repo.Update(ctx, card); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, card.ID)
	if got.Status != domain.StatusBlocked {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusBlocked)
	}
	if got.PINAttempts != 3 {
		t.Errorf("PINAttempts = %d, want 3", got.PINAttempts)
	}
}

func TestCardRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent card")
	}
}
