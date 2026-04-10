package template_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/infrastructure/postgres"
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

func TestTemplateRepo_CreateAndGetByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	tmpl, err := domain.NewTemplate("welcome_email", domain.ChannelEmail, "Welcome!", "Hello {{.Name}}", "en")
	if err != nil {
		t.Fatalf("creating template: %v", err)
	}

	if err := repo.Create(ctx, tmpl); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if tmpl.ID == "" {
		t.Fatal("expected non-empty ID after create")
	}

	found, err := repo.GetByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("repo.GetByID: %v", err)
	}
	if found.Slug != "welcome_email" {
		t.Errorf("slug = %q, want %q", found.Slug, "welcome_email")
	}
	if found.Version != 1 {
		t.Errorf("version = %d, want 1", found.Version)
	}
}

func TestTemplateRepo_GetBySlugAndLocale(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	tmpl, _ := domain.NewTemplate("otp_sms", domain.ChannelSMS, "", "OTP: {{.Code}}", "uz")
	tmpl.Activate()
	_ = repo.Create(ctx, tmpl)
	_ = repo.Update(ctx, tmpl) // persist status change

	found, err := repo.GetBySlugAndLocale(ctx, "otp_sms", "uz")
	if err != nil {
		t.Fatalf("repo.GetBySlugAndLocale: %v", err)
	}
	if found.Channel != domain.ChannelSMS {
		t.Errorf("channel = %q, want %q", found.Channel, domain.ChannelSMS)
	}
}

func TestTemplateRepo_ListByChannel(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	t1, _ := domain.NewTemplate("email1", domain.ChannelEmail, "S1", "B1", "en")
	t2, _ := domain.NewTemplate("sms1", domain.ChannelSMS, "", "B2", "en")
	t3, _ := domain.NewTemplate("email2", domain.ChannelEmail, "S2", "B3", "uz")
	_ = repo.Create(ctx, t1)
	_ = repo.Create(ctx, t2)
	_ = repo.Create(ctx, t3)

	emails, err := repo.ListByChannel(ctx, "EMAIL", 10, 0)
	if err != nil {
		t.Fatalf("repo.ListByChannel: %v", err)
	}
	if len(emails) != 2 {
		t.Errorf("email count = %d, want 2", len(emails))
	}

	all, err := repo.ListByChannel(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("repo.ListByChannel all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("all count = %d, want 3", len(all))
	}
}
