package iprule_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/infrastructure/postgres"
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

func TestIPRuleRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, err := domain.NewIPRule("192.168.1.100", domain.RuleTypeDeny, "Suspicious activity", "admin", nil)
	if err != nil {
		t.Fatalf("creating IP rule: %v", err)
	}

	if err := repo.Save(ctx, rule); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("expected rule ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.IPAddress != "192.168.1.100" {
		t.Errorf("ipAddress = %q, want %q", got.IPAddress, "192.168.1.100")
	}
	if got.RuleType != domain.RuleTypeDeny {
		t.Errorf("ruleType = %q, want %q", got.RuleType, domain.RuleTypeDeny)
	}
	if got.Reason != "Suspicious activity" {
		t.Errorf("reason = %q, want %q", got.Reason, "Suspicious activity")
	}
	if got.CreatedBy != "admin" {
		t.Errorf("createdBy = %q, want %q", got.CreatedBy, "admin")
	}
	if got.ExpiresAt != nil {
		t.Errorf("expiresAt = %v, want nil", got.ExpiresAt)
	}
}

func TestIPRuleRepo_SaveWithExpiration(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	expires := time.Now().Add(24 * time.Hour)
	rule, _ := domain.NewIPRule("10.0.0.50", domain.RuleTypeAllow, "Temporary whitelist", "ops", &expires)
	repo.Save(ctx, rule)

	got, err := repo.FindByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}
	if got.ExpiresAt == nil {
		t.Fatal("expected expiresAt to be set")
	}
	if got.RuleType != domain.RuleTypeAllow {
		t.Errorf("ruleType = %q, want %q", got.RuleType, domain.RuleTypeAllow)
	}
}

func TestIPRuleRepo_FindByIP(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, _ := domain.NewIPRule("172.16.0.1", domain.RuleTypeDeny, "Blocked", "security", nil)
	repo.Save(ctx, rule)

	got, err := repo.FindByIP(ctx, "172.16.0.1")
	if err != nil {
		t.Fatalf("repo.FindByIP: %v", err)
	}
	if got.ID != rule.ID {
		t.Errorf("ID = %q, want %q", got.ID, rule.ID)
	}
	if got.IPAddress != "172.16.0.1" {
		t.Errorf("ipAddress = %q, want %q", got.IPAddress, "172.16.0.1")
	}
}

func TestIPRuleRepo_ListAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	r1, _ := domain.NewIPRule("10.0.0.1", domain.RuleTypeAllow, "Office", "admin", nil)
	r2, _ := domain.NewIPRule("10.0.0.2", domain.RuleTypeDeny, "Blocked", "admin", nil)
	repo.Save(ctx, r1)
	repo.Save(ctx, r2)

	items, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("repo.ListAll: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
}

func TestIPRuleRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, _ := domain.NewIPRule("192.168.0.99", domain.RuleTypeDeny, "Temp block", "admin", nil)
	repo.Save(ctx, rule)

	if err := repo.Delete(ctx, rule.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, rule.ID)
	if err == nil {
		t.Fatal("expected error after deleting IP rule")
	}
}
