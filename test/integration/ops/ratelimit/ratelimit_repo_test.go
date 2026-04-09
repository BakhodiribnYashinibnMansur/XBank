package ratelimit_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/infrastructure/postgres"
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

func TestRateLimitRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, err := domain.NewRateLimitRule("/api/v1/transfers", 100, 60, "Transfer endpoint limit", true)
	if err != nil {
		t.Fatalf("creating rate limit rule: %v", err)
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

	if got.Key != "/api/v1/transfers" {
		t.Errorf("key = %q, want %q", got.Key, "/api/v1/transfers")
	}
	if got.MaxRequests != 100 {
		t.Errorf("maxRequests = %d, want 100", got.MaxRequests)
	}
	if got.WindowSeconds != 60 {
		t.Errorf("windowSeconds = %d, want 60", got.WindowSeconds)
	}
	if got.Description != "Transfer endpoint limit" {
		t.Errorf("description = %q, want %q", got.Description, "Transfer endpoint limit")
	}
	if !got.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestRateLimitRepo_FindByKey(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, _ := domain.NewRateLimitRule("/api/v1/auth/login", 10, 300, "Login rate limit", true)
	repo.Save(ctx, rule)

	got, err := repo.FindByKey(ctx, "/api/v1/auth/login")
	if err != nil {
		t.Fatalf("repo.FindByKey: %v", err)
	}
	if got.ID != rule.ID {
		t.Errorf("ID = %q, want %q", got.ID, rule.ID)
	}
	if got.MaxRequests != 10 {
		t.Errorf("maxRequests = %d, want 10", got.MaxRequests)
	}
}

func TestRateLimitRepo_FindAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	r1, _ := domain.NewRateLimitRule("/api/v1/accounts", 50, 60, "Accounts endpoint", true)
	r2, _ := domain.NewRateLimitRule("/api/v1/cards", 30, 60, "Cards endpoint", false)
	repo.Save(ctx, r1)
	repo.Save(ctx, r2)

	items, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("repo.FindAll: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
}

func TestRateLimitRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, _ := domain.NewRateLimitRule("/api/v1/payments", 200, 120, "Payments limit", true)
	repo.Save(ctx, rule)

	rule.Update(500, 300, "Updated payments limit", false)

	if err := repo.Update(ctx, rule); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, rule.ID)
	if got.MaxRequests != 500 {
		t.Errorf("maxRequests = %d, want 500", got.MaxRequests)
	}
	if got.WindowSeconds != 300 {
		t.Errorf("windowSeconds = %d, want 300", got.WindowSeconds)
	}
	if got.Description != "Updated payments limit" {
		t.Errorf("description = %q, want %q", got.Description, "Updated payments limit")
	}
	if got.Enabled {
		t.Error("expected Enabled to be false after update")
	}
}

func TestRateLimitRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rule, _ := domain.NewRateLimitRule("/api/v1/temp", 10, 60, "Temporary", true)
	repo.Save(ctx, rule)

	if err := repo.Delete(ctx, rule.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, rule.ID)
	if err == nil {
		t.Fatal("expected error after deleting rate limit rule")
	}
}
