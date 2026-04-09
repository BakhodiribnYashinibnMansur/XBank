package metric_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/infrastructure/postgres"
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

func TestMetricRepo_Save(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	m, err := domain.NewAppMetric("http_request_duration_seconds", 0.125, map[string]string{
		"method": "GET",
		"path":   "/api/v1/accounts",
		"status": "200",
	})
	if err != nil {
		t.Fatalf("creating metric: %v", err)
	}

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if m.ID == "" {
		t.Fatal("expected metric ID to be set after Save")
	}
}

func TestMetricRepo_FindByName(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	m1, _ := domain.NewAppMetric("db_query_duration", 0.05, map[string]string{"query": "SELECT"})
	m2, _ := domain.NewAppMetric("db_query_duration", 0.12, map[string]string{"query": "INSERT"})
	m3, _ := domain.NewAppMetric("http_request_total", 1.0, nil)
	repo.Save(ctx, m1)
	repo.Save(ctx, m2)
	repo.Save(ctx, m3)

	items, err := repo.FindByName(ctx, "db_query_duration")
	if err != nil {
		t.Fatalf("repo.FindByName: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	for _, item := range items {
		if item.Name != "db_query_duration" {
			t.Errorf("name = %q, want %q", item.Name, "db_query_duration")
		}
	}
}

func TestMetricRepo_FindByName_WithLabels(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	labels := map[string]string{"endpoint": "/health", "method": "GET"}
	m, _ := domain.NewAppMetric("api_latency", 0.003, labels)
	repo.Save(ctx, m)

	items, err := repo.FindByName(ctx, "api_latency")
	if err != nil {
		t.Fatalf("repo.FindByName: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}

	got := items[0]
	if got.Labels["endpoint"] != "/health" {
		t.Errorf("labels[endpoint] = %q, want %q", got.Labels["endpoint"], "/health")
	}
	if got.Labels["method"] != "GET" {
		t.Errorf("labels[method] = %q, want %q", got.Labels["method"], "GET")
	}
}

func TestMetricRepo_ListRecent(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		m, _ := domain.NewAppMetric(fmt.Sprintf("metric_%d", i), float64(i), nil)
		repo.Save(ctx, m)
	}

	items, err := repo.ListRecent(ctx, 3)
	if err != nil {
		t.Fatalf("repo.ListRecent: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
}

func TestMetricRepo_FindByName_Empty(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	items, err := repo.FindByName(ctx, "nonexistent_metric")
	if err != nil {
		t.Fatalf("repo.FindByName: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0", len(items))
	}
}
