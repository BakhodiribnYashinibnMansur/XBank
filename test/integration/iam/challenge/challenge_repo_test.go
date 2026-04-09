package challenge_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	challengedomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/domain"
	challengepg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/infrastructure/postgres"
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
	user, err := userdomain.NewUser("challenge-test@example.com", "$2a$12$hashedpassword", "Test", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestChallengeRepo_Create(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := challengepg.NewWriteRepo(pgc.Pool)

	ch, err := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "transfer", `{"amount":1000}`)
	if err != nil {
		t.Fatalf("NewChallenge: %v", err)
	}

	if err := repo.Create(ctx, ch); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if ch.ID == "" {
		t.Fatal("expected challenge ID to be set")
	}
}

func TestChallengeRepo_GetByID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := challengepg.NewWriteRepo(pgc.Pool)

	ch, _ := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "transfer", "meta")
	repo.Create(ctx, ch)

	got, err := repo.GetByID(ctx, ch.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.Action != "transfer" {
		t.Errorf("Action = %q, want %q", got.Action, "transfer")
	}
	if got.Status != challengedomain.StatusPending {
		t.Errorf("Status = %q, want %q", got.Status, challengedomain.StatusPending)
	}
}

func TestChallengeRepo_GetByToken(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := challengepg.NewWriteRepo(pgc.Pool)

	ch, _ := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "transfer", "meta")
	repo.Create(ctx, ch)

	// Verify the challenge to generate a token
	if err := ch.Verify(); err != nil {
		t.Fatalf("Challenge.Verify: %v", err)
	}
	if err := repo.Update(ctx, ch); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, err := repo.GetByToken(ctx, ch.Token)
	if err != nil {
		t.Fatalf("repo.GetByToken: %v", err)
	}
	if got.ID != ch.ID {
		t.Errorf("ID = %q, want %q", got.ID, ch.ID)
	}
	if got.Status != challengedomain.StatusVerified {
		t.Errorf("Status = %q, want %q", got.Status, challengedomain.StatusVerified)
	}
}

func TestChallengeRepo_Update(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := challengepg.NewWriteRepo(pgc.Pool)

	ch, _ := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "transfer", "meta")
	repo.Create(ctx, ch)

	ch.Fail()
	if err := repo.Update(ctx, ch); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, ch.ID)
	if got.Status != challengedomain.StatusFailed {
		t.Errorf("Status = %q, want %q", got.Status, challengedomain.StatusFailed)
	}
}

func TestChallengeRepo_CountPendingByUser(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := challengepg.NewWriteRepo(pgc.Pool)

	// No challenges yet
	count, err := repo.CountPendingByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountPendingByUser: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	// Create two pending challenges
	ch1, _ := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "transfer", "")
	repo.Create(ctx, ch1)

	ch2, _ := challengedomain.NewChallenge(userID, challengedomain.MethodPassword, "card_issue", "")
	repo.Create(ctx, ch2)

	count, err = repo.CountPendingByUser(ctx, userID)
	if err != nil {
		t.Fatalf("CountPendingByUser: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// Fail one — should only count 1 pending
	ch1.Fail()
	repo.Update(ctx, ch1)

	count, _ = repo.CountPendingByUser(ctx, userID)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
