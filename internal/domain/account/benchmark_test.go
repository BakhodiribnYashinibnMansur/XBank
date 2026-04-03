package account

import (
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

func BenchmarkAccount_Deposit(b *testing.B) {
	acc, _ := NewAccount("user-1", shared.UZS)
	amount := shared.Money{Amount: 10000, Currency: shared.UZS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc.Deposit(amount)
	}
}

func BenchmarkAccount_Withdraw(b *testing.B) {
	acc, _ := NewAccount("user-1", shared.UZS)
	// Pre-fund the account
	acc.Deposit(shared.Money{Amount: 1000000000, Currency: shared.UZS})

	amount := shared.Money{Amount: 100, Currency: shared.UZS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc.Withdraw(amount)
	}
}

func BenchmarkAccount_Apply_Events(b *testing.B) {
	events := make([]Event, 100)
	for i := 0; i < 100; i++ {
		if i == 0 {
			events[i] = Event{
				AggregateID: "acc-1",
				Type:        EventAccountOpened,
				Data:        AccountOpenedData{UserID: "u-1", AccountNumber: "1234", Currency: "UZS"},
				Version:     i + 1,
			}
		} else {
			events[i] = Event{
				AggregateID: "acc-1",
				Type:        EventCredited,
				Data:        CreditedData{Amount: 1000, Currency: "UZS"},
				Version:     i + 1,
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFromHistory(events)
	}
}

func BenchmarkAccount_LoadFromSnapshot(b *testing.B) {
	snap := SnapshotState{
		UserID:        "user-1",
		AccountNumber: "8600100000000001",
		Balance:       5000000,
		Currency:      "UZS",
		Status:        "ACTIVE",
	}

	// 10 events after snapshot
	events := make([]Event, 10)
	for i := 0; i < 10; i++ {
		events[i] = Event{
			AggregateID: "acc-1",
			Type:        EventCredited,
			Data:        CreditedData{Amount: 1000, Currency: "UZS"},
			Version:     101 + i,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFromSnapshot(snap, 100, events)
	}
}

func BenchmarkNewAccount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAccount("user-1", shared.UZS)
	}
}
