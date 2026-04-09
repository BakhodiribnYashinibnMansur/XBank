package domain

import "testing"

func TestCreatePair_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		transferID    string
		fromAccountID string
		toAccountID   string
		amount        int64
		currency      string
	}{
		{
			name:          "standard UZS transfer",
			transferID:    "tx-1",
			fromAccountID: "acc-from",
			toAccountID:   "acc-to",
			amount:        500000,
			currency:      "UZS",
		},
		{
			name:          "USD transfer",
			transferID:    "tx-2",
			fromAccountID: "acc-a",
			toAccountID:   "acc-b",
			amount:        10000,
			currency:      "USD",
		},
		{
			name:          "EUR transfer",
			transferID:    "tx-3",
			fromAccountID: "acc-x",
			toAccountID:   "acc-y",
			amount:        99999,
			currency:      "EUR",
		},
		{
			name:          "minimal amount",
			transferID:    "tx-4",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        1,
			currency:      "UZS",
		},
		{
			name:          "large amount",
			transferID:    "tx-5",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        999999999999,
			currency:      "UZS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debit, credit := CreatePair(tt.transferID, tt.fromAccountID, tt.toAccountID, tt.amount, tt.currency)

			// Debit entry
			if debit.AccountID != tt.fromAccountID {
				t.Errorf("debit AccountID: expected %s, got %s", tt.fromAccountID, debit.AccountID)
			}
			if debit.TransferID != tt.transferID {
				t.Errorf("debit TransferID: expected %s, got %s", tt.transferID, debit.TransferID)
			}
			if debit.EntryType != Debit {
				t.Errorf("debit EntryType: expected DEBIT, got %s", debit.EntryType)
			}
			if debit.Amount != tt.amount {
				t.Errorf("debit Amount: expected %d, got %d", tt.amount, debit.Amount)
			}
			if debit.Currency != tt.currency {
				t.Errorf("debit Currency: expected %s, got %s", tt.currency, debit.Currency)
			}

			// Credit entry
			if credit.AccountID != tt.toAccountID {
				t.Errorf("credit AccountID: expected %s, got %s", tt.toAccountID, credit.AccountID)
			}
			if credit.TransferID != tt.transferID {
				t.Errorf("credit TransferID: expected %s, got %s", tt.transferID, credit.TransferID)
			}
			if credit.EntryType != Credit {
				t.Errorf("credit EntryType: expected CREDIT, got %s", credit.EntryType)
			}
			if credit.Amount != tt.amount {
				t.Errorf("credit Amount: expected %d, got %d", tt.amount, credit.Amount)
			}
			if credit.Currency != tt.currency {
				t.Errorf("credit Currency: expected %s, got %s", tt.currency, credit.Currency)
			}

			// Symmetry: same amount, same timestamp
			if debit.Amount != credit.Amount {
				t.Error("debit and credit amounts must be equal")
			}
			if !debit.CreatedAt.Equal(credit.CreatedAt) {
				t.Error("debit and credit should have the same CreatedAt")
			}

			// Opposite types
			if debit.EntryType == credit.EntryType {
				t.Error("debit and credit should have opposite entry types")
			}

			// Different accounts
			if debit.AccountID == credit.AccountID {
				t.Error("debit and credit should be for different accounts")
			}
		})
	}
}

func TestCreatePair_DoubleEntryInvariant(t *testing.T) {
	// Double-entry bookkeeping invariant: for every transfer,
	// debit amount == credit amount (the books balance)
	amounts := []int64{1, 100, 50000, 1000000, 9999999999}

	for _, amount := range amounts {
		debit, credit := CreatePair("tx-1", "acc-1", "acc-2", amount, "UZS")

		if debit.Amount != credit.Amount {
			t.Errorf("double-entry violated: debit=%d, credit=%d for amount=%d",
				debit.Amount, credit.Amount, amount)
		}

		// Net effect should be zero (credit - debit = 0 for the system as a whole)
		netEffect := credit.Amount - debit.Amount
		if netEffect != 0 {
			t.Errorf("net effect should be 0, got %d", netEffect)
		}
	}
}

func TestEntry_FieldsAreSet(t *testing.T) {
	debit, credit := CreatePair("tx-abc", "from-acc", "to-acc", 42, "USD")

	if debit.CreatedAt.IsZero() {
		t.Error("debit CreatedAt should not be zero")
	}
	if credit.CreatedAt.IsZero() {
		t.Error("credit CreatedAt should not be zero")
	}

	// IDs are not set by CreatePair (they are set by the repository/DB)
	if debit.ID != "" {
		t.Error("debit ID should be empty (set by repository)")
	}
	if credit.ID != "" {
		t.Error("credit ID should be empty (set by repository)")
	}
}
