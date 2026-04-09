package user_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
)

var pgc *setup.PostgresContainer

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping integration tests (set INTEGRATION=1 to run)")
		os.Exit(0)
	}

	// Use a temporary *testing.T for setup
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

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	user, err := domain.NewUser("test@example.com", "$2a$12$hashedpassword", "John", "Doe")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if user.ID == "" {
		t.Fatal("expected user ID to be set after Create")
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}

	if got.Email != "test@example.com" {
		t.Errorf("email = %q, want %q", got.Email, "test@example.com")
	}
	if got.FirstName != "John" {
		t.Errorf("firstName = %q, want %q", got.FirstName, "John")
	}
	if got.LastName != "Doe" {
		t.Errorf("lastName = %q, want %q", got.LastName, "Doe")
	}
	if got.Role != domain.RoleCustomer {
		t.Errorf("role = %q, want %q", got.Role, domain.RoleCustomer)
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	user, _ := domain.NewUser("findme@example.com", "$2a$12$hash", "Jane", "Smith")
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	got, err := repo.GetByEmail(ctx, "findme@example.com")
	if err != nil {
		t.Fatalf("repo.GetByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("ID = %q, want %q", got.ID, user.ID)
	}
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
}

func TestUserRepo_ExistsByEmail(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	exists, err := repo.ExistsByEmail(ctx, "nobody@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail: %v", err)
	}
	if exists {
		t.Error("expected false for nonexistent email")
	}

	user, _ := domain.NewUser("somebody@example.com", "$2a$12$hash", "Bob", "")
	repo.Create(ctx, user)

	exists, err = repo.ExistsByEmail(ctx, "somebody@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail: %v", err)
	}
	if !exists {
		t.Error("expected true for existing email")
	}
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	user, _ := domain.NewUser("pwuser@example.com", "$2a$12$oldhash", "Alice", "Wonder")
	repo.Create(ctx, user)

	err := repo.UpdatePassword(ctx, user.ID, "$2a$12$newhash")
	if err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, _ := repo.GetByID(ctx, user.ID)
	if got.Password != "$2a$12$newhash" {
		t.Errorf("password = %q, want %q", got.Password, "$2a$12$newhash")
	}
}

func TestUserRepo_UpdateTOTP(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	user, _ := domain.NewUser("totp@example.com", "$2a$12$hash", "Charlie", "Brown")
	repo.Create(ctx, user)

	now := time.Now().Truncate(time.Microsecond)
	err := repo.UpdateTOTP(ctx, user.ID, "JBSWY3DPEHPK3PXP", true, &now)
	if err != nil {
		t.Fatalf("UpdateTOTP: %v", err)
	}

	got, _ := repo.GetByID(ctx, user.ID)
	if got.TOTPSecret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("totpSecret = %q, want %q", got.TOTPSecret, "JBSWY3DPEHPK3PXP")
	}
	if !got.TOTPEnabled {
		t.Error("expected TOTPEnabled to be true")
	}
}

func TestUserRepo_Anonymize(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	user, _ := domain.NewUser("gdpr@example.com", "$2a$12$hash", "David", "Delete")
	repo.Create(ctx, user)

	err := repo.Anonymize(ctx, user.ID)
	if err != nil {
		t.Fatalf("Anonymize: %v", err)
	}

	got, _ := repo.GetByID(ctx, user.ID)
	if got.FirstName != "[DELETED]" {
		t.Errorf("firstName = %q, want %q", got.FirstName, "[DELETED]")
	}
	if got.Password != "" {
		t.Errorf("password should be empty after anonymize, got %q", got.Password)
	}
}
