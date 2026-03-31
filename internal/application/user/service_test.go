package user

import (
	"context"
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mock"
)

func newTestService() (*Service, *mock.UserRepository) {
	repo := mock.NewUserRepository()
	svc := NewService(repo)
	return svc, repo
}

func TestRegister_Success(t *testing.T) {
	svc, _ := newTestService()

	u, err := svc.Register(context.Background(), "ali@example.com", "password123", "Ali", "Valiyev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID == "" {
		t.Error("ID should not be empty")
	}
	if u.Email != "ali@example.com" {
		t.Errorf("expected email: ali@example.com, got: %s", u.Email)
	}
	// password must be hashed, not plain text
	if u.Password == "password123" {
		t.Error("password should be hashed, not stored in plain text")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _ := newTestService()

	svc.Register(context.Background(), "ali@example.com", "password123", "Ali", "Valiyev")

	_, err := svc.Register(context.Background(), "ali@example.com", "password456", "Vali", "Aliyev")
	if err != user.ErrEmailExists {
		t.Errorf("expected: %v, got: %v", user.ErrEmailExists, err)
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.Register(context.Background(), "", "password123", "Ali", "Valiyev")
	if err != user.ErrInvalidEmail {
		t.Errorf("expected: %v, got: %v", user.ErrInvalidEmail, err)
	}
}

func TestGetByID_Success(t *testing.T) {
	svc, _ := newTestService()

	created, _ := svc.Register(context.Background(), "ali@example.com", "password123", "Ali", "Valiyev")

	found, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Email != "ali@example.com" {
		t.Errorf("expected: ali@example.com, got: %s", found.Email)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.GetByID(context.Background(), "non-existent-id")
	if err != user.ErrUserNotFound {
		t.Errorf("expected: %v, got: %v", user.ErrUserNotFound, err)
	}
}
