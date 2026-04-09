package domain

import (
	"testing"
)

func TestCreatePair(t *testing.T) {
	debit, credit := CreatePair("tx-1", "acc-from", "acc-to", 500000, "UZS")

	if debit == nil || credit == nil {
		t.Fatal("CreatePair should return two non-nil entries")
	}

	// Debit entry checks
	if debit.AccountID != "acc-from" {
		t.Errorf("debit AccountID expected acc-from, got: %s", debit.AccountID)
	}
	if debit.TransferID != "tx-1" {
		t.Errorf("debit TransferID expected tx-1, got: %s", debit.TransferID)
	}
	if debit.EntryType != Debit {
		t.Errorf("debit EntryType expected DEBIT, got: %s", debit.EntryType)
	}
	if debit.Amount != 500000 {
		t.Errorf("debit Amount expected 500000, got: %d", debit.Amount)
	}
	if debit.Currency != "UZS" {
		t.Errorf("debit Currency expected UZS, got: %s", debit.Currency)
	}

	// Credit entry checks
	if credit.AccountID != "acc-to" {
		t.Errorf("credit AccountID expected acc-to, got: %s", credit.AccountID)
	}
	if credit.TransferID != "tx-1" {
		t.Errorf("credit TransferID expected tx-1, got: %s", credit.TransferID)
	}
	if credit.EntryType != Credit {
		t.Errorf("credit EntryType expected CREDIT, got: %s", credit.EntryType)
	}
	if credit.Amount != 500000 {
		t.Errorf("credit Amount expected 500000, got: %d", credit.Amount)
	}

	// Both entries should have the same timestamp
	if !debit.CreatedAt.Equal(credit.CreatedAt) {
		t.Error("debit and credit should have the same CreatedAt timestamp")
	}
}

func TestCreatePair_DifferentCurrencies(t *testing.T) {
	tests := []struct {
		name     string
		currency string
	}{
		{"UZS currency", "UZS"},
		{"USD currency", "USD"},
		{"EUR currency", "EUR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debit, credit := CreatePair("tx-1", "acc-1", "acc-2", 100000, tt.currency)
			if debit.Currency != tt.currency {
				t.Errorf("debit Currency expected %s, got: %s", tt.currency, debit.Currency)
			}
			if credit.Currency != tt.currency {
				t.Errorf("credit Currency expected %s, got: %s", tt.currency, credit.Currency)
			}
		})
	}
}

func TestCreatePair_AmountSymmetry(t *testing.T) {
	amounts := []int64{1, 100, 999999, 1000000000}

	for _, amount := range amounts {
		debit, credit := CreatePair("tx-1", "acc-1", "acc-2", amount, "UZS")
		if debit.Amount != credit.Amount {
			t.Errorf("debit amount (%d) should equal credit amount (%d) for input %d",
				debit.Amount, credit.Amount, amount)
		}
	}
}

func TestEntryType_Constants(t *testing.T) {
	if Debit != "DEBIT" {
		t.Errorf("Debit constant expected DEBIT, got: %s", Debit)
	}
	if Credit != "CREDIT" {
		t.Errorf("Credit constant expected CREDIT, got: %s", Credit)
	}
}
