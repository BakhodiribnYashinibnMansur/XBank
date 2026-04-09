package systemerror_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
	"github.com/google/uuid"
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

func TestSystemErrorRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	sysErr, err := domain.NewSystemError("DB_CONN_FAILED", "Database connection timeout", "CRITICAL", "SYSTEM")
	if err != nil {
		t.Fatalf("creating system error: %v", err)
	}
	sysErr.ID = uuid.New().String()
	sysErr.WithContext("req-abc", "user-1", "192.168.1.1", "/api/accounts", "GET", "goroutine 1 ...", map[string]string{"db": "primary"})

	if err := repo.Save(ctx, sysErr); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}

	result, err := repo.FindByID(ctx, sysErr.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}
	got := result.(*domain.SystemError)

	if got.Code != "DB_CONN_FAILED" {
		t.Errorf("code = %q, want %q", got.Code, "DB_CONN_FAILED")
	}
	if got.Message != "Database connection timeout" {
		t.Errorf("message = %q, want %q", got.Message, "Database connection timeout")
	}
	if got.Severity != "CRITICAL" {
		t.Errorf("severity = %q, want %q", got.Severity, "CRITICAL")
	}
	if got.Category != "SYSTEM" {
		t.Errorf("category = %q, want %q", got.Category, "SYSTEM")
	}
	if got.RequestID != "req-abc" {
		t.Errorf("requestID = %q, want %q", got.RequestID, "req-abc")
	}
	if got.UserID != "user-1" {
		t.Errorf("userID = %q, want %q", got.UserID, "user-1")
	}
	if got.IPAddress != "192.168.1.1" {
		t.Errorf("ipAddress = %q, want %q", got.IPAddress, "192.168.1.1")
	}
	if got.Path != "/api/accounts" {
		t.Errorf("path = %q, want %q", got.Path, "/api/accounts")
	}
	if got.Method != "GET" {
		t.Errorf("method = %q, want %q", got.Method, "GET")
	}
	if got.Resolution != domain.StatusPending {
		t.Errorf("resolution = %q, want %q", got.Resolution, domain.StatusPending)
	}
	if got.Metadata["db"] != "primary" {
		t.Errorf("metadata[db] = %q, want %q", got.Metadata["db"], "primary")
	}
}

func TestSystemErrorRepo_Update_Resolve(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	sysErr, _ := domain.NewSystemError("TIMEOUT", "Request timeout", "HIGH", "NETWORK")
	sysErr.ID = uuid.New().String()
	repo.Save(ctx, sysErr)

	if err := sysErr.Resolve("admin-user"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if err := repo.Update(ctx, sysErr); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	result, _ := repo.FindByID(ctx, sysErr.ID)
	got := result.(*domain.SystemError)

	if got.Resolution != domain.StatusResolved {
		t.Errorf("resolution = %q, want %q", got.Resolution, domain.StatusResolved)
	}
	if got.ResolvedBy != "admin-user" {
		t.Errorf("resolvedBy = %q, want %q", got.ResolvedBy, "admin-user")
	}
	if got.ResolvedAt == nil {
		t.Fatal("expected ResolvedAt to be set")
	}
}
