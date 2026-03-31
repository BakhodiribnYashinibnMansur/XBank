package card

import "testing"

func TestNewCard(t *testing.T) {
	c, err := NewCard("acc-123", TypeDebit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusInactive {
		t.Errorf("new card should be INACTIVE, got: %s", c.Status)
	}
	if !ValidateLuhn(c.PAN) {
		t.Error("card PAN should pass Luhn check")
	}
	if c.MaskedPAN[:4] != "****" {
		t.Error("masked PAN should start with ****")
	}
}

func TestCard_Activate(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)

	if err := c.Activate("1234"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusActive {
		t.Errorf("expected ACTIVE, got: %s", c.Status)
	}
	if c.PINHash == "" {
		t.Error("PIN hash should not be empty")
	}
	if c.PINHash == "1234" {
		t.Error("PIN should be hashed, not plain text")
	}
}

func TestCard_Activate_InvalidPIN(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)

	err := c.Activate("12") // too short
	if err == nil {
		t.Error("expected error for short PIN")
	}
}

func TestCard_VerifyPIN(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)
	c.Activate("1234")

	if err := c.VerifyPIN("1234"); err != nil {
		t.Errorf("correct PIN should pass: %v", err)
	}

	if err := c.VerifyPIN("0000"); err == nil {
		t.Error("wrong PIN should fail")
	}
}

func TestCard_PINBruteForce(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)
	c.Activate("1234")

	// 3 wrong attempts → blocked
	c.VerifyPIN("0000")
	c.VerifyPIN("0000")
	err := c.VerifyPIN("0000")

	if err != ErrPINAttemptsExceeded {
		t.Errorf("expected PIN_LOCKED, got: %v", err)
	}
	if c.Status != StatusBlocked {
		t.Errorf("card should be BLOCKED, got: %s", c.Status)
	}

	// Even correct PIN should fail now
	if err := c.VerifyPIN("1234"); err == nil {
		t.Error("blocked card should reject all PINs")
	}
}

func TestCard_ChangePIN(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)
	c.Activate("1234")

	if err := c.ChangePIN("1234", "5678"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Old PIN should fail
	if err := c.VerifyPIN("1234"); err == nil {
		t.Error("old PIN should fail after change")
	}

	// Reset attempts for new PIN test
	c.PINAttempts = 0

	// New PIN should work
	if err := c.VerifyPIN("5678"); err != nil {
		t.Errorf("new PIN should work: %v", err)
	}
}

func TestCard_Block_Unblock(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)
	c.Activate("1234")

	c.Block()
	if c.Status != StatusBlocked {
		t.Error("should be BLOCKED")
	}

	c.Unblock()
	if c.Status != StatusActive {
		t.Error("should be ACTIVE")
	}
	if c.PINAttempts != 0 {
		t.Error("PIN attempts should be reset after unblock")
	}
}

func TestCard_Cancel(t *testing.T) {
	c, _ := NewCard("acc-123", TypeDebit)
	c.Activate("1234")

	c.Cancel()
	if c.Status != StatusCancelled {
		t.Error("should be CANCELLED")
	}
}
