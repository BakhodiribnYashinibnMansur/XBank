package translation_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/infrastructure/postgres"
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

func TestTranslationRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	tr, err := domain.NewTranslation("error.not_found", domain.LangEn, "Resource not found", "errors")
	if err != nil {
		t.Fatalf("creating translation: %v", err)
	}

	if err := repo.Save(ctx, tr); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if tr.ID == "" {
		t.Fatal("expected translation ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, tr.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.Key != "error.not_found" {
		t.Errorf("key = %q, want %q", got.Key, "error.not_found")
	}
	if got.Language != domain.LangEn {
		t.Errorf("language = %q, want %q", got.Language, domain.LangEn)
	}
	if got.Value != "Resource not found" {
		t.Errorf("value = %q, want %q", got.Value, "Resource not found")
	}
	if got.Group != "errors" {
		t.Errorf("group = %q, want %q", got.Group, "errors")
	}
}

func TestTranslationRepo_FindByKeyAndLanguage(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	trEn, _ := domain.NewTranslation("greeting", domain.LangEn, "Hello", "auth")
	trUz, _ := domain.NewTranslation("greeting", domain.LangUz, "Salom", "auth")
	repo.Save(ctx, trEn)
	repo.Save(ctx, trUz)

	got, err := repo.FindByKeyAndLanguage(ctx, "greeting", domain.LangUz)
	if err != nil {
		t.Fatalf("repo.FindByKeyAndLanguage: %v", err)
	}
	if got.Value != "Salom" {
		t.Errorf("value = %q, want %q", got.Value, "Salom")
	}
	if got.Language != domain.LangUz {
		t.Errorf("language = %q, want %q", got.Language, domain.LangUz)
	}
}

func TestTranslationRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	tr, _ := domain.NewTranslation("btn.submit", domain.LangRu, "Otpravit", "dashboard")
	repo.Save(ctx, tr)

	if err := tr.Update("Otpravit'"); err != nil {
		t.Fatalf("tr.Update: %v", err)
	}

	if err := repo.Update(ctx, tr); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, tr.ID)
	if got.Value != "Otpravit'" {
		t.Errorf("value = %q, want %q", got.Value, "Otpravit'")
	}
}

func TestTranslationRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	tr, _ := domain.NewTranslation("temp.key", domain.LangEn, "Temporary", "misc")
	repo.Save(ctx, tr)

	if err := repo.Delete(ctx, tr.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, tr.ID)
	if err == nil {
		t.Fatal("expected error after deleting translation")
	}
}
