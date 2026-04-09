package exchange_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange/infrastructure/postgres"
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

func TestExchangeRepo_UpsertAndGetRate(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rate := &domain.Rate{
		FromCurrency: "USD",
		ToCurrency:   "UZS",
		BuyRate:      1260000,
		SellRate:     1265050,
	}

	if err := repo.Upsert(ctx, rate); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if rate.ID == "" {
		t.Fatal("expected rate ID to be set after upsert")
	}

	got, err := repo.GetRate(ctx, "USD", "UZS")
	if err != nil {
		t.Fatalf("GetRate: %v", err)
	}
	if got.BuyRate != 1260000 {
		t.Errorf("BuyRate = %d, want 1260000", got.BuyRate)
	}
	if got.SellRate != 1265050 {
		t.Errorf("SellRate = %d, want 1265050", got.SellRate)
	}
}

func TestExchangeRepo_UpsertUpdatesExisting(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rate := &domain.Rate{
		FromCurrency: "EUR",
		ToCurrency:   "UZS",
		BuyRate:      1350000,
		SellRate:     1360000,
	}
	if err := repo.Upsert(ctx, rate); err != nil {
		t.Fatalf("first Upsert: %v", err)
	}
	firstID := rate.ID

	// Upsert again with new rates
	rate2 := &domain.Rate{
		FromCurrency: "EUR",
		ToCurrency:   "UZS",
		BuyRate:      1370000,
		SellRate:     1380000,
	}
	if err := repo.Upsert(ctx, rate2); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	if rate2.ID != firstID {
		t.Errorf("expected same ID after upsert, got %q and %q", firstID, rate2.ID)
	}

	got, _ := repo.GetRate(ctx, "EUR", "UZS")
	if got.SellRate != 1380000 {
		t.Errorf("SellRate = %d, want 1380000 after upsert", got.SellRate)
	}
}

func TestExchangeRepo_GetRate_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	_, err := repo.GetRate(ctx, "XXX", "YYY")
	if err == nil {
		t.Fatal("expected error for nonexistent rate")
	}
}

func TestExchangeRepo_ListAll(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	rates := []*domain.Rate{
		{FromCurrency: "EUR", ToCurrency: "UZS", BuyRate: 1350000, SellRate: 1360000},
		{FromCurrency: "USD", ToCurrency: "UZS", BuyRate: 1260000, SellRate: 1265050},
	}
	for _, r := range rates {
		if err := repo.Upsert(ctx, r); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("got %d rates, want 2", len(all))
	}
}
