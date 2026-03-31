package shared

import "testing"

func TestNewMoney(t *testing.T) {
	m, err := NewMoney(1500050, UZS) // 15000.50 UZS
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if m.Amount != 1500050 {
		t.Errorf("Amount kutilgan: 1500050, kelgan: %d", m.Amount)
	}
	if m.String() != "15000.50 UZS" {
		t.Errorf("String kutilgan: 15000.50 UZS, kelgan: %s", m.String())
	}
}

func TestNewMoney_Negative(t *testing.T) {
	_, err := NewMoney(-100, UZS)
	if err != ErrNegativeAmount {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrNegativeAmount, err)
	}
}

func TestMoney_Add(t *testing.T) {
	a, _ := NewMoney(10000, UZS) // 100.00
	b, _ := NewMoney(5000, UZS)  // 50.00

	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if result.Amount != 15000 {
		t.Errorf("Kutilgan: 15000, kelgan: %d", result.Amount)
	}
}

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	a, _ := NewMoney(10000, UZS)
	b, _ := NewMoney(5000, USD)

	_, err := a.Add(b)
	if err != ErrCurrencyMismatch {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrCurrencyMismatch, err)
	}
}

func TestMoney_Subtract(t *testing.T) {
	a, _ := NewMoney(10000, UZS)
	b, _ := NewMoney(3000, UZS)

	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if result.Amount != 7000 {
		t.Errorf("Kutilgan: 7000, kelgan: %d", result.Amount)
	}
}

func TestMoney_Subtract_InsufficientFunds(t *testing.T) {
	a, _ := NewMoney(3000, UZS)
	b, _ := NewMoney(10000, UZS)

	_, err := a.Subtract(b)
	if err != ErrInsufficientFunds {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrInsufficientFunds, err)
	}
}
