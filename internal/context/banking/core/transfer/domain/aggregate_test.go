package domain

import (
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

func TestNewTransfer_Success(t *testing.T) {
	amount, _ := shared.NewMoney(500000, shared.UZS)
	tr, err := NewTransfer("acc-1", "acc-2", amount, "Test o'tkazma")
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if tr.Status != StatusPending {
		t.Errorf("Status PENDING bo'lishi kerak, kelgan: %s", tr.Status)
	}
	if tr.FromAccountID != "acc-1" {
		t.Errorf("FromAccountID kutilgan: acc-1, kelgan: %s", tr.FromAccountID)
	}
}

func TestNewTransfer_SameAccount(t *testing.T) {
	amount, _ := shared.NewMoney(500000, shared.UZS)
	_, err := NewTransfer("acc-1", "acc-1", amount, "")
	if err != ErrSameAccount {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrSameAccount, err)
	}
}

func TestNewTransfer_ZeroAmount(t *testing.T) {
	amount := shared.Money{Amount: 0, Currency: shared.UZS}
	_, err := NewTransfer("acc-1", "acc-2", amount, "")
	if err != ErrInvalidAmount {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrInvalidAmount, err)
	}
}

func TestTransfer_Complete(t *testing.T) {
	amount, _ := shared.NewMoney(500000, shared.UZS)
	tr, _ := NewTransfer("acc-1", "acc-2", amount, "")
	tr.Complete()
	if tr.Status != StatusCompleted {
		t.Errorf("Status COMPLETED bo'lishi kerak, kelgan: %s", tr.Status)
	}
}

func TestTransfer_Fail(t *testing.T) {
	amount, _ := shared.NewMoney(500000, shared.UZS)
	tr, _ := NewTransfer("acc-1", "acc-2", amount, "")
	tr.Fail("mablag' yetarli emas")
	if tr.Status != StatusFailed {
		t.Errorf("Status FAILED bo'lishi kerak, kelgan: %s", tr.Status)
	}
	if tr.FailureReason != "mablag' yetarli emas" {
		t.Errorf("FailureReason kutilgan: mablag' yetarli emas, kelgan: %s", tr.FailureReason)
	}
}
