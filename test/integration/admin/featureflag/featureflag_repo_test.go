package featureflag_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/infrastructure/postgres"
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

func TestFeatureFlagRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	flag, err := domain.NewFeatureFlag("dark_mode", "Enable dark mode UI", domain.FlagTypeBool, "false")
	if err != nil {
		t.Fatalf("creating feature flag: %v", err)
	}

	if err := repo.Save(ctx, flag); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if flag.ID == "" {
		t.Fatal("expected flag ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, flag.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.Key != "dark_mode" {
		t.Errorf("key = %q, want %q", got.Key, "dark_mode")
	}
	if got.Description != "Enable dark mode UI" {
		t.Errorf("description = %q, want %q", got.Description, "Enable dark mode UI")
	}
	if got.FlagType != domain.FlagTypeBool {
		t.Errorf("flagType = %q, want %q", got.FlagType, domain.FlagTypeBool)
	}
	if got.DefaultValue != "false" {
		t.Errorf("defaultValue = %q, want %q", got.DefaultValue, "false")
	}
	if got.Active {
		t.Error("expected Active to be false")
	}
	if got.RolloutPct != 0 {
		t.Errorf("rolloutPct = %d, want 0", got.RolloutPct)
	}
}

func TestFeatureFlagRepo_FindByKey(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	flag, _ := domain.NewFeatureFlag("beta_access", "Beta feature access", domain.FlagTypeString, "disabled")
	if err := repo.Save(ctx, flag); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}

	got, err := repo.FindByKey(ctx, "beta_access")
	if err != nil {
		t.Fatalf("repo.FindByKey: %v", err)
	}
	if got.ID != flag.ID {
		t.Errorf("ID = %q, want %q", got.ID, flag.ID)
	}
	if got.Key != "beta_access" {
		t.Errorf("key = %q, want %q", got.Key, "beta_access")
	}
}

func TestFeatureFlagRepo_FindAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	f1, _ := domain.NewFeatureFlag("aaa_flag", "First flag", domain.FlagTypeBool, "true")
	f2, _ := domain.NewFeatureFlag("zzz_flag", "Second flag", domain.FlagTypeInt, "42")
	repo.Save(ctx, f1)
	repo.Save(ctx, f2)

	flags, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("repo.FindAll: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("len(flags) = %d, want 2", len(flags))
	}
	// Should be ordered by key
	if flags[0].Key != "aaa_flag" {
		t.Errorf("first flag key = %q, want %q", flags[0].Key, "aaa_flag")
	}
	if flags[1].Key != "zzz_flag" {
		t.Errorf("second flag key = %q, want %q", flags[1].Key, "zzz_flag")
	}
}

func TestFeatureFlagRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	flag, _ := domain.NewFeatureFlag("update_me", "Original desc", domain.FlagTypeBool, "false")
	repo.Save(ctx, flag)

	newDesc := "Updated description"
	newActive := true
	newPct := 50
	if err := flag.Update(&newDesc, nil, &newActive, &newPct); err != nil {
		t.Fatalf("flag.Update: %v", err)
	}

	if err := repo.Update(ctx, flag); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, flag.ID)
	if got.Description != "Updated description" {
		t.Errorf("description = %q, want %q", got.Description, "Updated description")
	}
	if !got.Active {
		t.Error("expected Active to be true after update")
	}
	if got.RolloutPct != 50 {
		t.Errorf("rolloutPct = %d, want 50", got.RolloutPct)
	}
}

func TestFeatureFlagRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	flag, _ := domain.NewFeatureFlag("delete_me", "To be deleted", domain.FlagTypeBool, "false")
	repo.Save(ctx, flag)

	if err := repo.Delete(ctx, flag.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, flag.ID)
	if err == nil {
		t.Fatal("expected error after deleting flag")
	}
}
