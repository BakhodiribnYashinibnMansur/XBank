package reconciliation_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/infrastructure/postgres"
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
	user, err := userdomain.NewUser("recon@bank.com", "$2a$12$hash", "Recon", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestReconciliationRepo_Save(t *testing.T) {
	truncate(t)
	repo := pg.NewRunRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	run := &domain.ReconciliationRun{
		ID:           "a0000000-0000-4000-8000-000000000001",
		UserID:       userID,
		TotalChecked: 100,
		Matches:      98,
		Mismatches:   2,
		Status:       "COMPLETED",
		CreatedAt:    time.Now(),
	}

	if err := repo.Save(ctx, run); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func TestReconciliationRepo_SaveMultiple(t *testing.T) {
	truncate(t)
	repo := pg.NewRunRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	for i := 0; i < 3; i++ {
		run := &domain.ReconciliationRun{
			ID:           fmt.Sprintf("a0000000-0000-4000-8000-00000000000%d", i+1),
			UserID:       userID,
			TotalChecked: 50 * (i + 1),
			Matches:      50*(i+1) - i,
			Mismatches:   i,
			Status:       "COMPLETED",
			CreatedAt:    time.Now(),
		}
		if err := repo.Save(ctx, run); err != nil {
			t.Fatalf("Save run %d: %v", i, err)
		}
	}
}

func TestReconciliationRepo_SavePartialFailure(t *testing.T) {
	truncate(t)
	repo := pg.NewRunRepo(pgc.Pool)
	ctx := context.Background()
	userID := createTestUser(t)

	run := &domain.ReconciliationRun{
		ID:           "a0000000-0000-4000-8000-000000000010",
		UserID:       userID,
		TotalChecked: 200,
		Matches:      150,
		Mismatches:   50,
		Status:       "PARTIAL_FAILURE",
		CreatedAt:    time.Now(),
	}

	if err := repo.Save(ctx, run); err != nil {
		t.Fatalf("Save partial failure: %v", err)
	}
}
