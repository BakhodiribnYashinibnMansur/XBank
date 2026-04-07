package domain

import "testing"

func TestNewMoney(t *testing.T) {
	m, err := NewMoney(1500050, UZS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Amount != 1500050 {
		t.Errorf("expected 1500050, got %d", m.Amount)
	}
	if m.Currency != UZS {
		t.Errorf("expected UZS, got %s", m.Currency)
	}
}

func TestNewMoney_Negative(t *testing.T) {
	_, err := NewMoney(-100, UZS)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestMoney_Add(t *testing.T) {
	a, _ := NewMoney(100, UZS)
	b, _ := NewMoney(200, UZS)
	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Amount != 300 {
		t.Errorf("expected 300, got %d", result.Amount)
	}
}

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	a, _ := NewMoney(100, UZS)
	b, _ := NewMoney(200, USD)
	_, err := a.Add(b)
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}
}

func TestMoney_Subtract(t *testing.T) {
	a, _ := NewMoney(500, UZS)
	b, _ := NewMoney(200, UZS)
	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Amount != 300 {
		t.Errorf("expected 300, got %d", result.Amount)
	}
}

func TestMoney_Subtract_InsufficientFunds(t *testing.T) {
	a, _ := NewMoney(100, UZS)
	b, _ := NewMoney(200, UZS)
	_, err := a.Subtract(b)
	if err == nil {
		t.Fatal("expected insufficient funds error")
	}
}
