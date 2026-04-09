package audit_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	auditdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/domain"
	auditpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/infrastructure/postgres"
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

func TestAuditRepo_CreateAuditLog(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := auditpg.NewWriteRepo(pgc.Pool)

	log, err := auditdomain.NewAuditLog(
		"User", "user-123", "LOGIN", "actor-1",
		map[string]any{"ip": "10.0.0.1"},
		"192.168.1.1", "Mozilla/5.0",
	)
	if err != nil {
		t.Fatalf("NewAuditLog: %v", err)
	}

	if err := repo.CreateAuditLog(ctx, log); err != nil {
		t.Fatalf("repo.CreateAuditLog: %v", err)
	}
	if log.ID == "" {
		t.Fatal("expected audit log ID to be set")
	}
}

func TestAuditRepo_ListAuditLogs(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := auditpg.NewWriteRepo(pgc.Pool)

	// Create multiple logs
	l1, _ := auditdomain.NewAuditLog("User", "user-1", "LOGIN", "actor-1", nil, "10.0.0.1", "Chrome")
	repo.CreateAuditLog(ctx, l1)

	l2, _ := auditdomain.NewAuditLog("User", "user-1", "LOGOUT", "actor-1", nil, "10.0.0.1", "Chrome")
	repo.CreateAuditLog(ctx, l2)

	l3, _ := auditdomain.NewAuditLog("Account", "acc-1", "TRANSFER", "actor-2", nil, "10.0.0.2", "Firefox")
	repo.CreateAuditLog(ctx, l3)

	// List all
	logs, total, err := repo.ListAuditLogs(ctx, auditdomain.AuditFilter{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("repo.ListAuditLogs: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(logs) != 3 {
		t.Errorf("len(logs) = %d, want 3", len(logs))
	}

	// Filter by aggregate type
	logs, total, err = repo.ListAuditLogs(ctx, auditdomain.AuditFilter{
		AggregateType: "User",
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("repo.ListAuditLogs (filtered): %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(logs) != 2 {
		t.Errorf("len(logs) = %d, want 2", len(logs))
	}

	// Filter by action
	logs, total, err = repo.ListAuditLogs(ctx, auditdomain.AuditFilter{
		Action: "LOGIN",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("repo.ListAuditLogs (action filter): %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
}

func TestAuditRepo_CreateEndpointHistory(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := auditpg.NewWriteRepo(pgc.Pool)

	h, err := auditdomain.NewEndpointHistory("GET", "/api/v1/accounts", 200, "user-1", "10.0.0.1", 42)
	if err != nil {
		t.Fatalf("NewEndpointHistory: %v", err)
	}

	if err := repo.CreateEndpointHistory(ctx, h); err != nil {
		t.Fatalf("repo.CreateEndpointHistory: %v", err)
	}
	if h.ID == "" {
		t.Fatal("expected endpoint history ID to be set")
	}
}

func TestAuditRepo_ListEndpointHistory(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := auditpg.NewWriteRepo(pgc.Pool)

	h1, _ := auditdomain.NewEndpointHistory("GET", "/api/v1/accounts", 200, "user-1", "10.0.0.1", 30)
	repo.CreateEndpointHistory(ctx, h1)

	h2, _ := auditdomain.NewEndpointHistory("POST", "/api/v1/transfers", 201, "user-1", "10.0.0.1", 120)
	repo.CreateEndpointHistory(ctx, h2)

	h3, _ := auditdomain.NewEndpointHistory("GET", "/api/v1/accounts", 200, "user-2", "10.0.0.2", 25)
	repo.CreateEndpointHistory(ctx, h3)

	// List all
	entries, total, err := repo.ListEndpointHistory(ctx, auditdomain.EndpointFilter{
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("repo.ListEndpointHistory: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}

	// Filter by user
	entries, total, err = repo.ListEndpointHistory(ctx, auditdomain.EndpointFilter{
		UserID: "user-1",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("repo.ListEndpointHistory (user filter): %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}

	// Filter by method
	entries, total, err = repo.ListEndpointHistory(ctx, auditdomain.EndpointFilter{
		Method: "POST",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("repo.ListEndpointHistory (method filter): %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(entries) != 1 {
		t.Errorf("len(entries) = %d, want 1", len(entries))
	}
}
