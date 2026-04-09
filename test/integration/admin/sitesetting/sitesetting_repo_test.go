package sitesetting_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/infrastructure/postgres"
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

func TestSiteSettingRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	setting, err := domain.NewSiteSetting("app.name", "XBank", domain.SettingTypeGeneral, "Application name")
	if err != nil {
		t.Fatalf("creating site setting: %v", err)
	}

	if err := repo.Save(ctx, setting); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if setting.ID == "" {
		t.Fatal("expected setting ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, setting.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.Key != "app.name" {
		t.Errorf("key = %q, want %q", got.Key, "app.name")
	}
	if got.Value != "XBank" {
		t.Errorf("value = %q, want %q", got.Value, "XBank")
	}
	if got.SettingType != domain.SettingTypeGeneral {
		t.Errorf("settingType = %q, want %q", got.SettingType, domain.SettingTypeGeneral)
	}
	if got.Description != "Application name" {
		t.Errorf("description = %q, want %q", got.Description, "Application name")
	}
}

func TestSiteSettingRepo_FindByKey(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	setting, _ := domain.NewSiteSetting("smtp.host", "mail.example.com", domain.SettingTypeEmail, "SMTP server")
	repo.Save(ctx, setting)

	got, err := repo.FindByKey(ctx, "smtp.host")
	if err != nil {
		t.Fatalf("repo.FindByKey: %v", err)
	}
	if got.ID != setting.ID {
		t.Errorf("ID = %q, want %q", got.ID, setting.ID)
	}
	if got.Value != "mail.example.com" {
		t.Errorf("value = %q, want %q", got.Value, "mail.example.com")
	}
}

func TestSiteSettingRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	setting, _ := domain.NewSiteSetting("max.retries", "3", domain.SettingTypeSecurity, "Max login retries")
	repo.Save(ctx, setting)

	newValue := "5"
	newDesc := "Updated max retries"
	setting.Update(&newValue, &newDesc)

	if err := repo.Update(ctx, setting); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, setting.ID)
	if got.Value != "5" {
		t.Errorf("value = %q, want %q", got.Value, "5")
	}
	if got.Description != "Updated max retries" {
		t.Errorf("description = %q, want %q", got.Description, "Updated max retries")
	}
}

func TestSiteSettingRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	setting, _ := domain.NewSiteSetting("temp.key", "temp_value", domain.SettingTypeGeneral, "Temporary")
	repo.Save(ctx, setting)

	if err := repo.Delete(ctx, setting.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, setting.ID)
	if err == nil {
		t.Fatal("expected error after deleting setting")
	}
}
