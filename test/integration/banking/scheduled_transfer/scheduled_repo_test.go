package scheduled_transfer_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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
	user, err := userdomain.NewUser("sched@bank.com", "$2a$12$hash", "Sched", "User")
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

func TestScheduledRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewScheduledRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	amount := kerneldomain.Money{Amount: 50000, Currency: kerneldomain.UZS}
	executeAt := time.Now().Add(24 * time.Hour)

	st, err := domain.NewScheduledTransfer(userID, fromAccID, toAccID, amount, "scheduled test", executeAt)
	if err != nil {
		t.Fatalf("NewScheduledTransfer: %v", err)
	}

	if err := repo.Create(ctx, st); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	got, err := repo.GetByID(ctx, st.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.FromAccountID != fromAccID {
		t.Errorf("FromAccountID = %q, want %q", got.FromAccountID, fromAccID)
	}
	if got.Amount.Amount != 50000 {
		t.Errorf("Amount = %d, want 50000", got.Amount.Amount)
	}
	if got.Status != domain.ScheduledPending {
		t.Errorf("Status = %q, want %q", got.Status, domain.ScheduledPending)
	}
	if got.Description != "scheduled test" {
		t.Errorf("Description = %q, want %q", got.Description, "scheduled test")
	}
}

func TestScheduledRepo_ListByUserID(t *testing.T) {
	truncate(t)
	repo := pg.NewScheduledRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	for i := 0; i < 3; i++ {
		amount := kerneldomain.Money{Amount: int64(1000 * (i + 1)), Currency: kerneldomain.UZS}
		executeAt := time.Now().Add(time.Duration(i+1) * 24 * time.Hour)
		st, _ := domain.NewScheduledTransfer(userID, fromAccID, toAccID, amount, fmt.Sprintf("sched %d", i), executeAt)
		if err := repo.Create(ctx, st); err != nil {
			t.Fatalf("creating scheduled transfer %d: %v", i, err)
		}
	}

	results, total, err := repo.ListByUserID(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
}

func TestScheduledRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewScheduledRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	amount := kerneldomain.Money{Amount: 30000, Currency: kerneldomain.UZS}
	executeAt := time.Now().Add(24 * time.Hour)
	st, _ := domain.NewScheduledTransfer(userID, fromAccID, toAccID, amount, "update test", executeAt)
	repo.Create(ctx, st)

	transferUUID := "e0000000-0000-4000-8000-000000000123"
	st.MarkExecuted(transferUUID)
	if err := repo.Update(ctx, st); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, st.ID)
	if got.Status != domain.ScheduledExecuted {
		t.Errorf("Status = %q, want %q", got.Status, domain.ScheduledExecuted)
	}
	if got.TransferID != transferUUID {
		t.Errorf("TransferID = %q, want %q", got.TransferID, transferUUID)
	}
	if got.ExecutedAt == nil {
		t.Error("expected ExecutedAt to be set")
	}
}

func TestScheduledRepo_FetchDue(t *testing.T) {
	truncate(t)
	repo := pg.NewScheduledRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)
	fromAccID := createTestAccount(t, userID)
	toAccID := createTestAccount(t, userID)

	// Create a scheduled transfer with execute_at in the past (already due)
	amount := kerneldomain.Money{Amount: 10000, Currency: kerneldomain.UZS}
	st := &domain.ScheduledTransfer{
		ID:            "d0000000-0000-4000-8000-000000000001",
		UserID:        userID,
		FromAccountID: fromAccID,
		ToAccountID:   toAccID,
		Amount:        amount,
		Description:   "due transfer",
		Status:        domain.ScheduledPending,
		ExecuteAt:     time.Now().Add(-1 * time.Hour), // in the past
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, st); err != nil {
		t.Fatalf("Create due transfer: %v", err)
	}

	// Create a future scheduled transfer (not due)
	stFuture := &domain.ScheduledTransfer{
		ID:            "d0000000-0000-4000-8000-000000000002",
		UserID:        userID,
		FromAccountID: fromAccID,
		ToAccountID:   toAccID,
		Amount:        amount,
		Description:   "future transfer",
		Status:        domain.ScheduledPending,
		ExecuteAt:     time.Now().Add(24 * time.Hour), // in the future
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, stFuture); err != nil {
		t.Fatalf("Create future transfer: %v", err)
	}

	due, err := repo.FetchDue(ctx, 10)
	if err != nil {
		t.Fatalf("FetchDue: %v", err)
	}
	if len(due) != 1 {
		t.Errorf("got %d due transfers, want 1", len(due))
	}
	if len(due) > 0 && due[0].ID != st.ID {
		t.Errorf("due transfer ID = %q, want %q", due[0].ID, st.ID)
	}
}

func TestScheduledRepo_GetByID_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewScheduledRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent scheduled transfer")
	}
}
