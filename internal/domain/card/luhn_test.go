package card

import "testing"

func TestValidateLuhn(t *testing.T) {
	// Valid card numbers
	validNumbers := []string{
		"4539148803436467",
		"4916338506082832",
		"5425233430109903",
	}
	for _, n := range validNumbers {
		if !ValidateLuhn(n) {
			t.Errorf("expected %s to be valid", n)
		}
	}

	// Invalid card numbers
	invalidNumbers := []string{
		"4539148803436468",
		"1234567890123456",
	}
	for _, n := range invalidNumbers {
		if ValidateLuhn(n) {
			t.Errorf("expected %s to be invalid", n)
		}
	}
}

func TestGenerateCardNumber(t *testing.T) {
	pan, err := GenerateCardNumber()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pan) != 16 {
		t.Errorf("expected 16 digits, got: %d", len(pan))
	}

	// Must start with BIN 486486
	if pan[:6] != "486486" {
		t.Errorf("expected BIN 486486, got: %s", pan[:6])
	}

	// Must pass Luhn check
	if !ValidateLuhn(pan) {
		t.Errorf("generated number %s fails Luhn check", pan)
	}
}

func TestGenerateCardNumber_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pan, _ := GenerateCardNumber()
		if seen[pan] {
			t.Errorf("duplicate card number: %s", pan)
		}
		seen[pan] = true
	}
}

func TestMaskPAN(t *testing.T) {
	masked := MaskPAN("4861234567891234")
	if masked != "**** **** **** 1234" {
		t.Errorf("expected: **** **** **** 1234, got: %s", masked)
	}
}
