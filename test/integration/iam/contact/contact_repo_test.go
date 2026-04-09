package contact_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	contactdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/contact/domain"
	contactpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/contact/infrastructure/postgres"
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

var userCounter int

func createTestUser(t *testing.T, label string) string {
	t.Helper()
	userCounter++
	userRepo := userpg.NewWriteRepo(pgc.Pool)
	email := fmt.Sprintf("contact-%s-%d@example.com", label, userCounter)
	user, err := userdomain.NewUser(email, "$2a$12$hashedpassword", "Test", label)
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestContactRepo_Add(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	contactID := createTestUser(t, "friend")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	contact, err := contactdomain.NewContact(ownerID, contactID, "My Friend")
	if err != nil {
		t.Fatalf("NewContact: %v", err)
	}

	if err := repo.Add(ctx, contact); err != nil {
		t.Fatalf("repo.Add: %v", err)
	}
}

func TestContactRepo_Add_DuplicateNoError(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	contactID := createTestUser(t, "friend")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	c1, _ := contactdomain.NewContact(ownerID, contactID, "Friend v1")
	repo.Add(ctx, c1)

	// Adding again should not error (ON CONFLICT DO NOTHING)
	c2, _ := contactdomain.NewContact(ownerID, contactID, "Friend v2")
	if err := repo.Add(ctx, c2); err != nil {
		t.Fatalf("repo.Add (duplicate): %v", err)
	}
}

func TestContactRepo_ListByOwnerID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	friend1 := createTestUser(t, "friend1")
	friend2 := createTestUser(t, "friend2")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	c1, _ := contactdomain.NewContact(ownerID, friend1, "First")
	repo.Add(ctx, c1)
	c2, _ := contactdomain.NewContact(ownerID, friend2, "Second")
	repo.Add(ctx, c2)

	contacts, err := repo.ListByOwnerID(ctx, ownerID, 10, 0)
	if err != nil {
		t.Fatalf("repo.ListByOwnerID: %v", err)
	}
	if len(contacts) != 2 {
		t.Errorf("len(contacts) = %d, want 2", len(contacts))
	}
}

func TestContactRepo_CountByOwnerID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	friend1 := createTestUser(t, "friend1")
	friend2 := createTestUser(t, "friend2")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	c1, _ := contactdomain.NewContact(ownerID, friend1, "A")
	repo.Add(ctx, c1)
	c2, _ := contactdomain.NewContact(ownerID, friend2, "B")
	repo.Add(ctx, c2)

	count, err := repo.CountByOwnerID(ctx, ownerID)
	if err != nil {
		t.Fatalf("repo.CountByOwnerID: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestContactRepo_Delete(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	friendID := createTestUser(t, "friend")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	contact, _ := contactdomain.NewContact(ownerID, friendID, "To Delete")
	repo.Add(ctx, contact)

	if err := repo.Delete(ctx, ownerID, friendID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	exists, _ := repo.IsContact(ctx, ownerID, friendID)
	if exists {
		t.Error("expected contact to be deleted")
	}
}

func TestContactRepo_IsContact(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	friendID := createTestUser(t, "friend")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	// Before adding
	exists, err := repo.IsContact(ctx, ownerID, friendID)
	if err != nil {
		t.Fatalf("repo.IsContact: %v", err)
	}
	if exists {
		t.Error("expected false before adding contact")
	}

	// After adding
	contact, _ := contactdomain.NewContact(ownerID, friendID, "My Friend")
	repo.Add(ctx, contact)

	exists, err = repo.IsContact(ctx, ownerID, friendID)
	if err != nil {
		t.Fatalf("repo.IsContact: %v", err)
	}
	if !exists {
		t.Error("expected true after adding contact")
	}
}

func TestContactRepo_GetByID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	ownerID := createTestUser(t, "owner")
	friendID := createTestUser(t, "friend")
	repo := contactpg.NewWriteRepo(pgc.Pool)

	contact, _ := contactdomain.NewContact(ownerID, friendID, "Buddy")
	repo.Add(ctx, contact)

	// Get the ID by listing
	contacts, _ := repo.ListByOwnerID(ctx, ownerID, 10, 0)
	if len(contacts) == 0 {
		t.Fatal("expected at least one contact")
	}

	got, err := repo.GetByID(ctx, contacts[0].ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if got.OwnerID != ownerID {
		t.Errorf("OwnerID = %q, want %q", got.OwnerID, ownerID)
	}
	if got.CustomName != "Buddy" {
		t.Errorf("CustomName = %q, want %q", got.CustomName, "Buddy")
	}
}
