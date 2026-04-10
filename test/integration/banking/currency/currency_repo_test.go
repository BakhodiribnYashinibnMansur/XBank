package currency_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/infrastructure/postgres"
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

func TestCurrencyRepo_CreateAndGetByCode(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	cur, err := domain.NewCurrency("UZS", "Uzbek Sum", "so'm", 0)
	if err != nil {
		t.Fatalf("creating currency: %v", err)
	}

	if err := repo.Create(ctx, cur); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}
	if cur.ID == "" {
		t.Fatal("expected non-empty ID after create")
	}

	found, err := repo.GetByCode(ctx, "UZS")
	if err != nil {
		t.Fatalf("repo.GetByCode: %v", err)
	}
	if found.Name != "Uzbek Sum" {
		t.Errorf("name = %q, want %q", found.Name, "Uzbek Sum")
	}
}

func TestCurrencyRepo_ListAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	usd, _ := domain.NewCurrency("USD", "US Dollar", "$", 2)
	uzs, _ := domain.NewCurrency("UZS", "Uzbek Sum", "so'm", 0)
	_ = repo.Create(ctx, usd)
	_ = repo.Create(ctx, uzs)

	list, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("repo.ListAll: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("count = %d, want 2", len(list))
	}
}

func TestCurrencyRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	cur, _ := domain.NewCurrency("EUR", "Euro", "E", 2)
	_ = repo.Create(ctx, cur)

	cur.Symbol = "€"
	cur.Deactivate()
	if err := repo.Update(ctx, cur); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	updated, _ := repo.GetByID(ctx, cur.ID)
	if updated.Symbol != "€" {
		t.Errorf("symbol = %q, want %q", updated.Symbol, "€")
	}
	if updated.Status != domain.StatusInactive {
		t.Errorf("status = %q, want %q", updated.Status, domain.StatusInactive)
	}
}
