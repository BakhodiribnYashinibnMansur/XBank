package healthcheck_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/infrastructure/postgres"
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

func TestHealthcheckRepo_SaveAndGetLatest(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	health := domain.NewSystemHealth([]domain.ComponentCheck{
		{Name: "postgres", Status: domain.StatusHealthy},
		{Name: "redis", Status: domain.StatusDegraded, Message: "slow"},
	})

	record := &domain.HealthRecord{
		Status:     health.Status,
		Components: `[{"name":"postgres","status":"HEALTHY"},{"name":"redis","status":"DEGRADED","message":"slow"}]`,
		CheckedAt:  health.CheckedAt,
	}

	if err := repo.Save(ctx, record); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if record.ID == "" {
		t.Fatal("expected non-empty ID after save")
	}

	latest, err := repo.GetLatest(ctx)
	if err != nil {
		t.Fatalf("repo.GetLatest: %v", err)
	}
	if latest.Status != domain.StatusDegraded {
		t.Errorf("status = %q, want %q", latest.Status, domain.StatusDegraded)
	}
}

func TestHealthcheckRepo_ListHistory(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		record := &domain.HealthRecord{
			Status:     domain.StatusHealthy,
			Components: `[]`,
			CheckedAt:  domain.NewSystemHealth(nil).CheckedAt,
		}
		_ = repo.Save(ctx, record)
	}

	records, err := repo.ListHistory(ctx, 3, 0)
	if err != nil {
		t.Fatalf("repo.ListHistory: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("count = %d, want 3 (limit=3)", len(records))
	}

	count, err := repo.CountHistory(ctx)
	if err != nil {
		t.Fatalf("repo.CountHistory: %v", err)
	}
	if count != 5 {
		t.Errorf("total = %d, want 5", count)
	}
}
