package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// skipMigrations — migrations that cannot run in testcontainers:
//   - 015: RLS + FORCE policies require app.current_user_id setting
//   - 020: pg_cron extension not available in standard postgres image
var skipMigrations = map[string]bool{
	"015_enable_rls.sql":    true,
	"020_setup_pg_cron.sql": true,
}

// PostgresContainer wraps a testcontainers PostgreSQL instance.
type PostgresContainer struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
	DSN       string
}

// MustPostgres starts a PostgreSQL container, applies migrations, and returns a pool.
// Call Teardown() to clean up — typically in TestMain via defer.
func MustPostgres(t *testing.T) *PostgresContainer {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("xbank_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("starting postgres container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("getting connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("creating pool: %v", err)
	}

	pc := &PostgresContainer{
		Container: container,
		Pool:      pool,
		DSN:       dsn,
	}

	applyMigrations(t, pool, ctx)

	return pc
}

// Teardown closes the pool and terminates the container.
func (pc *PostgresContainer) Teardown() {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		pc.Container.Terminate(context.Background())
	}
}

// applyMigrations reads SQL migration files and executes them in order.
// Skips migrations listed in skipMigrations.
func applyMigrations(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
	t.Helper()

	migrationsDir := findMigrationsDir(t)
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("reading migrations dir: %v", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if skipMigrations[e.Name()] {
			t.Logf("skipping migration: %s", e.Name())
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			t.Fatalf("reading migration %s: %v", f, err)
		}

		upSQL := extractGooseUp(string(content))
		if upSQL == "" {
			continue
		}

		if _, err := pool.Exec(ctx, upSQL); err != nil {
			t.Fatalf("applying migration %s: %v", f, err)
		}
	}
	t.Logf("applied %d migrations", len(files))
}

// extractGooseUp extracts the SQL between "-- +goose Up" and "-- +goose Down".
func extractGooseUp(content string) string {
	upIdx := strings.Index(content, "-- +goose Up")
	if upIdx < 0 {
		return ""
	}

	sql := content[upIdx+len("-- +goose Up"):]

	downIdx := strings.Index(sql, "-- +goose Down")
	if downIdx >= 0 {
		sql = sql[:downIdx]
	}

	return strings.TrimSpace(sql)
}

// findMigrationsDir walks up from the current directory to find migrations/.
func findMigrationsDir(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working dir: %v", err)
	}

	for {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find migrations directory")
		}
		dir = parent
	}
}

// TruncateAll truncates all user-created tables for test isolation.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename NOT LIKE 'goose_%'
		ORDER BY tablename`)
	if err != nil {
		t.Fatalf("listing tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scanning table name: %v", err)
		}
		tables = append(tables, name)
	}

	if len(tables) > 0 {
		_, err := pool.Exec(ctx,
			fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", ")))
		if err != nil {
			t.Fatalf("truncating tables: %v", err)
		}
	}
}
