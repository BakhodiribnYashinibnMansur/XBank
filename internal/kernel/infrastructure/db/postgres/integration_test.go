// Integration tests for PostgreSQL repositories.
// These tests require a running PostgreSQL instance.
// Set DATABASE_URL env or they will be skipped.
//
// Run: DATABASE_URL=postgres://postgres:postgres@localhost:5432/xbank_test?sslmode=disable go test ./internal/infrastructure/postgres/ -v -run Integration
package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	user "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func init() {
	logger.Init(true)
	metrics.Register()
}

func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Cannot connect to DB: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("DB ping failed: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// cleanupUser removes test user by email
func cleanupUser(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
}

// ── User Repository Integration Tests ────────────────

func TestIntegration_UserRepository_CreateAndGetByEmail(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_create@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	u := &user.User{
		Email:     testEmail,
		Password:  "$2a$12$fakehashfakehashfakehashfakehashfakehashfakehashfake",
		FirstName: "Test",
		LastName:  "User",
		Role:      user.RoleCustomer,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create
	err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if u.ID == "" {
		t.Fatal("User ID should be set after create")
	}

	// GetByEmail
	found, err := repo.GetByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if found.ID != u.ID {
		t.Errorf("ID mismatch: got %s, want %s", found.ID, u.ID)
	}
	if found.Email != testEmail {
		t.Errorf("Email mismatch: got %s, want %s", found.Email, testEmail)
	}
	if found.FirstName != "Test" {
		t.Errorf("FirstName mismatch: got %s, want Test", found.FirstName)
	}
	if found.Role != user.RoleCustomer {
		t.Errorf("Role mismatch: got %s, want CUSTOMER", found.Role)
	}
}

func TestIntegration_UserRepository_GetByID(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_getbyid@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	u := &user.User{
		Email:     testEmail,
		Password:  "$2a$12$fakehashfakehashfakehashfakehashfakehashfakehashfake",
		FirstName: "GetByID",
		LastName:  "Test",
		Role:      user.RoleCustomer,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.Create(ctx, u)

	found, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.Email != testEmail {
		t.Errorf("Email mismatch: got %s", found.Email)
	}
}

func TestIntegration_UserRepository_ExistsByEmail(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_exists@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	// Before create
	exists, err := repo.ExistsByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}
	if exists {
		t.Error("Should not exist before create")
	}

	// After create
	now := time.Now()
	repo.Create(ctx, &user.User{
		Email: testEmail, Password: "hash", FirstName: "X", LastName: "Y",
		Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
	})

	exists, err = repo.ExistsByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}
	if !exists {
		t.Error("Should exist after create")
	}
}

func TestIntegration_UserRepository_UpdatePassword(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_updatepw@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	u := &user.User{
		Email: testEmail, Password: "old_hash", FirstName: "PW", LastName: "Test",
		Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
	}
	repo.Create(ctx, u)

	err := repo.UpdatePassword(ctx, u.ID, "new_hash")
	if err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	found, _ := repo.GetByID(ctx, u.ID)
	if found.Password != "new_hash" {
		t.Errorf("Password not updated: got %s", found.Password)
	}
}

func TestIntegration_UserRepository_DuplicateEmail(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_dup@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	u1 := &user.User{
		Email: testEmail, Password: "hash1", FirstName: "First", LastName: "User",
		Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
	}
	repo.Create(ctx, u1)

	u2 := &user.User{
		Email: testEmail, Password: "hash2", FirstName: "Second", LastName: "User",
		Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
	}
	err := repo.Create(ctx, u2)
	if err == nil {
		t.Error("Duplicate email should fail")
	}
}

func TestIntegration_UserRepository_GetByID_NotFound(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-999999999999")
	if err == nil {
		t.Error("Non-existent user should return error")
	}
}

func TestIntegration_UserRepository_Anonymize(t *testing.T) {
	pool := getTestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_anon@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM users WHERE first_name = '[DELETED]' AND id = $1", "")
	})

	now := time.Now()
	u := &user.User{
		Email: testEmail, Password: "hash", FirstName: "Anon", LastName: "Test",
		Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
	}
	repo.Create(ctx, u)

	err := repo.Anonymize(ctx, u.ID)
	if err != nil {
		t.Fatalf("Anonymize failed: %v", err)
	}

	found, _ := repo.GetByID(ctx, u.ID)
	if found.FirstName != "[DELETED]" {
		t.Errorf("FirstName should be [DELETED], got %s", found.FirstName)
	}
	if found.LastName != "[DELETED]" {
		t.Errorf("LastName should be [DELETED], got %s", found.LastName)
	}
	if found.Password != "" {
		t.Errorf("Password should be empty after anonymize, got %s", found.Password)
	}

	// Cleanup anonymized user
	pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
}

// ── Transaction Manager Integration Tests ────────────

func TestIntegration_TxManager_WithTx(t *testing.T) {
	pool := getTestPool(t)
	txManager := NewTxManager(pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_tx@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	err := txManager.WithTx(ctx, func(txCtx context.Context) error {
		u := &user.User{
			Email: testEmail, Password: "hash", FirstName: "TX", LastName: "Test",
			Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
		}
		return repo.Create(txCtx, u)
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	exists, _ := repo.ExistsByEmail(ctx, testEmail)
	if !exists {
		t.Error("User should exist after committed transaction")
	}
}

func TestIntegration_TxManager_Rollback(t *testing.T) {
	pool := getTestPool(t)
	txManager := NewTxManager(pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	testEmail := "integration_test_rollback@xbank.test"
	cleanupUser(t, pool, testEmail)
	t.Cleanup(func() { cleanupUser(t, pool, testEmail) })

	now := time.Now()
	err := txManager.WithTx(ctx, func(txCtx context.Context) error {
		u := &user.User{
			Email: testEmail, Password: "hash", FirstName: "Rollback", LastName: "Test",
			Role: user.RoleCustomer, CreatedAt: now, UpdatedAt: now,
		}
		repo.Create(txCtx, u)
		return os.ErrNotExist // force rollback
	})
	if err == nil {
		t.Fatal("WithTx should return error")
	}

	exists, _ := repo.ExistsByEmail(ctx, testEmail)
	if exists {
		t.Error("User should NOT exist after rolled back transaction")
	}
}

func TestIntegration_TxManager_Serializable(t *testing.T) {
	pool := getTestPool(t)
	txManager := NewTxManager(pool)
	ctx := context.Background()

	// Just verify serializable transaction works without error
	err := txManager.WithSerializableTx(ctx, func(txCtx context.Context) error {
		var count int
		return pool.QueryRow(txCtx, "SELECT 1").Scan(&count)
	})
	if err != nil {
		t.Fatalf("WithSerializableTx failed: %v", err)
	}
}
