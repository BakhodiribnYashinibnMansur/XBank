package statistics_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/statistics/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/statistics/infrastructure/postgres"
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

func TestSnapshotRepo_Save(t *testing.T) {
	truncate(t)
	repo := pg.NewSnapshotRepo(pgc.Pool)
	ctx := context.Background()

	today := time.Now().Truncate(24 * time.Hour)
	snapshot := &domain.DailySnapshot{
		ID:             uuid.New().String(),
		Date:           today,
		TotalUsers:     1000,
		TotalAccounts:  2500,
		ActiveAccounts: 1800,
		TotalTransfers: 15000,
		TotalCards:     900,
		PendingKYC:     42,
		FlaggedFraud:   3,
		SystemErrors:   7,
	}

	if err := repo.Save(ctx, snapshot); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}

	// Verify by querying directly
	var totalUsers int64
	err := pgc.Pool.QueryRow(ctx,
		`SELECT total_users FROM statistics_snapshots WHERE id = $1`, snapshot.ID,
	).Scan(&totalUsers)
	if err != nil {
		t.Fatalf("querying snapshot: %v", err)
	}
	if totalUsers != 1000 {
		t.Errorf("totalUsers = %d, want 1000", totalUsers)
	}
}

func TestSnapshotRepo_Upsert(t *testing.T) {
	truncate(t)
	repo := pg.NewSnapshotRepo(pgc.Pool)
	ctx := context.Background()

	today := time.Now().Truncate(24 * time.Hour)

	// First insert
	s1 := &domain.DailySnapshot{
		ID:             uuid.New().String(),
		Date:           today,
		TotalUsers:     100,
		TotalAccounts:  200,
		ActiveAccounts: 150,
		TotalTransfers: 500,
		TotalCards:     50,
		PendingKYC:     10,
		FlaggedFraud:   1,
		SystemErrors:   2,
	}
	if err := repo.Save(ctx, s1); err != nil {
		t.Fatalf("repo.Save (first): %v", err)
	}

	// Upsert with same date, different values
	s2 := &domain.DailySnapshot{
		ID:             uuid.New().String(),
		Date:           today,
		TotalUsers:     200,
		TotalAccounts:  400,
		ActiveAccounts: 300,
		TotalTransfers: 1000,
		TotalCards:     100,
		PendingKYC:     20,
		FlaggedFraud:   2,
		SystemErrors:   5,
	}
	if err := repo.Save(ctx, s2); err != nil {
		t.Fatalf("repo.Save (upsert): %v", err)
	}

	// Should have only one row for today
	var count int
	pgc.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM statistics_snapshots WHERE date = $1`, today,
	).Scan(&count)
	if count != 1 {
		t.Errorf("count = %d, want 1 (upsert should not create duplicates)", count)
	}

	// Values should be updated
	var totalUsers int64
	pgc.Pool.QueryRow(ctx,
		`SELECT total_users FROM statistics_snapshots WHERE date = $1`, today,
	).Scan(&totalUsers)
	if totalUsers != 200 {
		t.Errorf("totalUsers = %d, want 200 (upserted value)", totalUsers)
	}
}
