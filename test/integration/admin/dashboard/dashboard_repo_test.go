package dashboard_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/infrastructure/postgres"
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

func TestDashboardRepo_GetOverview(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	stats, err := repo.GetOverview(ctx)
	if err != nil {
		t.Fatalf("repo.GetOverview: %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	// On empty DB, all counts should be 0
	if stats.TotalUsers != 0 {
		t.Errorf("TotalUsers = %d, want 0 on empty DB", stats.TotalUsers)
	}
}
