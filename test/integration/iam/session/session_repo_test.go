package session_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	sessiondomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/domain"
	sessionpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/infrastructure/postgres"
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
	user, err := userdomain.NewUser("session-test@example.com", "$2a$12$hashedpassword", "Test", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestSessionRepo_Create(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := sessionpg.NewWriteRepo(pgc.Pool)

	session, err := sessiondomain.NewSession(
		userID,
		"sha256-refresh-token-hash",
		"Mozilla/5.0",
		"192.168.1.1",
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if session.ID == "" {
		t.Fatal("expected session ID to be set after Create")
	}
}

func TestSessionRepo_GetByRefreshToken(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := sessionpg.NewWriteRepo(pgc.Pool)

	tokenHash := "sha256-unique-token-hash"
	session, _ := sessiondomain.NewSession(userID, tokenHash, "Chrome/120", "10.0.0.1", time.Now().Add(24*time.Hour))
	repo.Create(ctx, session)

	got, err := repo.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		t.Fatalf("repo.GetByRefreshToken: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("ID = %q, want %q", got.ID, session.ID)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
	if got.UserAgent != "Chrome/120" {
		t.Errorf("UserAgent = %q, want %q", got.UserAgent, "Chrome/120")
	}
}

func TestSessionRepo_GetByRefreshToken_NotFound(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := sessionpg.NewWriteRepo(pgc.Pool)

	_, err := repo.GetByRefreshToken(ctx, "nonexistent-token")
	if err == nil {
		t.Fatal("expected error for nonexistent token")
	}
}

func TestSessionRepo_DeleteByID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := sessionpg.NewWriteRepo(pgc.Pool)

	session, _ := sessiondomain.NewSession(userID, "token-to-delete", "Safari", "172.16.0.1", time.Now().Add(24*time.Hour))
	repo.Create(ctx, session)

	if err := repo.DeleteByID(ctx, session.ID); err != nil {
		t.Fatalf("repo.DeleteByID: %v", err)
	}

	_, err := repo.GetByRefreshToken(ctx, "token-to-delete")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestSessionRepo_DeleteAllByUserID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := sessionpg.NewWriteRepo(pgc.Pool)

	// Create multiple sessions
	s1, _ := sessiondomain.NewSession(userID, "token-1", "Chrome", "10.0.0.1", time.Now().Add(24*time.Hour))
	repo.Create(ctx, s1)

	s2, _ := sessiondomain.NewSession(userID, "token-2", "Firefox", "10.0.0.2", time.Now().Add(24*time.Hour))
	repo.Create(ctx, s2)

	if err := repo.DeleteAllByUserID(ctx, userID); err != nil {
		t.Fatalf("repo.DeleteAllByUserID: %v", err)
	}

	// Both sessions should be gone
	_, err := repo.GetByRefreshToken(ctx, "token-1")
	if err == nil {
		t.Error("expected error for token-1 after DeleteAll")
	}
	_, err = repo.GetByRefreshToken(ctx, "token-2")
	if err == nil {
		t.Error("expected error for token-2 after DeleteAll")
	}
}
