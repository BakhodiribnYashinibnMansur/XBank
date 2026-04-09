package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/infrastructure/postgres"
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

func TestIntegrationRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	integ, err := domain.NewIntegration(
		"PaymentGateway",
		"https://api.gateway.com",
		"sk_test_abc123",
		domain.StatusActive,
		"https://xbank.com/webhooks/gateway",
	)
	if err != nil {
		t.Fatalf("creating integration: %v", err)
	}

	if err := repo.Save(ctx, integ); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if integ.ID == "" {
		t.Fatal("expected integration ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, integ.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.Name != "PaymentGateway" {
		t.Errorf("name = %q, want %q", got.Name, "PaymentGateway")
	}
	if got.BaseURL != "https://api.gateway.com" {
		t.Errorf("baseURL = %q, want %q", got.BaseURL, "https://api.gateway.com")
	}
	if got.APIKey != "sk_test_abc123" {
		t.Errorf("apiKey = %q, want %q", got.APIKey, "sk_test_abc123")
	}
	if got.Status != domain.StatusActive {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusActive)
	}
	if got.WebhookURL != "https://xbank.com/webhooks/gateway" {
		t.Errorf("webhookURL = %q, want %q", got.WebhookURL, "https://xbank.com/webhooks/gateway")
	}
}

func TestIntegrationRepo_FindByName(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	integ, _ := domain.NewIntegration("SMSProvider", "https://sms.api.com", "key_sms_123", domain.StatusActive, "")
	repo.Save(ctx, integ)

	got, err := repo.FindByName(ctx, "SMSProvider")
	if err != nil {
		t.Fatalf("repo.FindByName: %v", err)
	}
	if got.ID != integ.ID {
		t.Errorf("ID = %q, want %q", got.ID, integ.ID)
	}
	if got.BaseURL != "https://sms.api.com" {
		t.Errorf("baseURL = %q, want %q", got.BaseURL, "https://sms.api.com")
	}
}

func TestIntegrationRepo_ListAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	i1, _ := domain.NewIntegration("ServiceA", "https://a.com", "key_a", domain.StatusActive, "")
	i2, _ := domain.NewIntegration("ServiceB", "https://b.com", "key_b", domain.StatusInactive, "")
	repo.Save(ctx, i1)
	repo.Save(ctx, i2)

	items, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("repo.ListAll: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
}

func TestIntegrationRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	integ, _ := domain.NewIntegration("OldService", "https://old.com", "old_key", domain.StatusActive, "")
	repo.Save(ctx, integ)

	integ.Update("https://new.com", "new_key", domain.StatusSuspended, "https://xbank.com/webhooks/new")

	if err := repo.Update(ctx, integ); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, integ.ID)
	if got.BaseURL != "https://new.com" {
		t.Errorf("baseURL = %q, want %q", got.BaseURL, "https://new.com")
	}
	if got.APIKey != "new_key" {
		t.Errorf("apiKey = %q, want %q", got.APIKey, "new_key")
	}
	if got.Status != domain.StatusSuspended {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusSuspended)
	}
	if got.WebhookURL != "https://xbank.com/webhooks/new" {
		t.Errorf("webhookURL = %q, want %q", got.WebhookURL, "https://xbank.com/webhooks/new")
	}
}

func TestIntegrationRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	integ, _ := domain.NewIntegration("TempService", "https://temp.com", "temp_key", domain.StatusInactive, "")
	repo.Save(ctx, integ)

	if err := repo.Delete(ctx, integ.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, integ.ID)
	if err == nil {
		t.Fatal("expected error after deleting integration")
	}
}
