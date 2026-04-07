package domain

import (
	"testing"
	"time"
)

func TestNewHold(t *testing.T) {
	h, err := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if h.Status != HoldStatusHeld {
		t.Errorf("expected HELD, got %s", h.Status)
	}
	if h.Amount != 5000 {
		t.Errorf("expected 5000, got %d", h.Amount)
	}
	if h.ExpiresAt.Before(time.Now()) {
		t.Error("expires_at should be in the future")
	}
}

func TestNewHold_ZeroAmount(t *testing.T) {
	_, err := NewHold("card-1", "acc-1", "Shop", 0, "USD")
	if err == nil {
		t.Error("should reject zero amount")
	}
}

func TestNewHold_NegativeAmount(t *testing.T) {
	_, err := NewHold("card-1", "acc-1", "Shop", -100, "USD")
	if err == nil {
		t.Error("should reject negative amount")
	}
}

func TestHold_Capture(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	if err := h.Capture(); err != nil {
		t.Fatal(err)
	}
	if h.Status != HoldStatusCaptured {
		t.Errorf("expected CAPTURED, got %s", h.Status)
	}
	if h.CapturedAt == nil {
		t.Error("captured_at should be set")
	}
}

func TestHold_Capture_AlreadyCaptured(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	h.Capture()
	if err := h.Capture(); err == nil {
		t.Error("should reject double capture")
	}
}

func TestHold_Release(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	if err := h.Release(); err != nil {
		t.Fatal(err)
	}
	if h.Status != HoldStatusReleased {
		t.Errorf("expected RELEASED, got %s", h.Status)
	}
	if h.ReleasedAt == nil {
		t.Error("released_at should be set")
	}
}

func TestHold_Release_AlreadyReleased(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	h.Release()
	if err := h.Release(); err == nil {
		t.Error("should reject double release")
	}
}

func TestHold_Capture_Expired(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	h.ExpiresAt = time.Now().Add(-1 * time.Second) // force expire

	if err := h.Capture(); err == nil {
		t.Error("should reject capture on expired hold")
	}
}

func TestHold_IsExpired(t *testing.T) {
	h, _ := NewHold("card-1", "acc-1", "Amazon", 5000, "USD")
	if h.IsExpired() {
		t.Error("new hold should not be expired")
	}

	h.ExpiresAt = time.Now().Add(-1 * time.Second)
	if !h.IsExpired() {
		t.Error("hold with past expiry should be expired")
	}
}
