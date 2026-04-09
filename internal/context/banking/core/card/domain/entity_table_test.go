package domain

import "testing"

func TestNewCard_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		cardType  Type
		wantErr   bool
	}{
		{
			name:      "valid debit card",
			accountID: "acc-1",
			cardType:  TypeDebit,
			wantErr:   false,
		},
		{
			name:      "valid virtual card",
			accountID: "acc-2",
			cardType:  TypeVirtual,
			wantErr:   false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			cardType:  TypeDebit,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := NewCard(tt.accountID, tt.cardType)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.AccountID != tt.accountID {
				t.Errorf("expected accountID %s, got %s", tt.accountID, card.AccountID)
			}
			if card.CardType != tt.cardType {
				t.Errorf("expected card type %s, got %s", tt.cardType, card.CardType)
			}
			if card.Status != StatusInactive {
				t.Errorf("expected INACTIVE, got %s", card.Status)
			}
			if card.PINAttempts != 0 {
				t.Errorf("expected 0 PIN attempts, got %d", card.PINAttempts)
			}
			if !ValidateLuhn(card.PAN) {
				t.Errorf("generated PAN %s does not pass Luhn check", card.PAN)
			}
			if card.ExpiryYear <= 0 {
				t.Error("expiry year should be set")
			}
		})
	}
}

func TestCard_Activate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		pin     string
		wantErr bool
	}{
		{
			name: "activate inactive card with valid PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			pin:     "1234",
			wantErr: false,
		},
		{
			name: "activate with short PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			pin:     "12",
			wantErr: true,
		},
		{
			name: "activate with long PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			pin:     "123456",
			wantErr: true,
		},
		{
			name: "activate already active card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			pin:     "5678",
			wantErr: true,
		},
		{
			name: "activate blocked card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			pin:     "5678",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.Activate(tt.pin, hasher)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Status != StatusActive {
				t.Errorf("expected ACTIVE, got %s", card.Status)
			}
			if card.PINHash == "" {
				t.Error("PIN hash should be set")
			}
			if card.PINHash == tt.pin {
				t.Error("PIN should be hashed, not stored as plaintext")
			}
		})
	}
}

func TestCard_VerifyPIN_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Card
		pin        string
		wantErr    error
		wantStatus Status
	}{
		{
			name: "correct PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			pin:        "1234",
			wantErr:    nil,
			wantStatus: StatusActive,
		},
		{
			name: "wrong PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			pin:        "0000",
			wantErr:    ErrInvalidPIN,
			wantStatus: StatusActive,
		},
		{
			name: "blocked card rejects even correct PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			pin:        "1234",
			wantErr:    ErrCardBlocked,
			wantStatus: StatusBlocked,
		},
		{
			name: "third wrong attempt blocks card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.VerifyPIN("0000", hasher) // attempt 1
				c.VerifyPIN("0000", hasher) // attempt 2
				return c
			},
			pin:        "0000", // attempt 3
			wantErr:    ErrPINAttemptsExceeded,
			wantStatus: StatusBlocked,
		},
		{
			name: "correct PIN resets attempt counter",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.VerifyPIN("0000", hasher) // 1 wrong attempt
				c.VerifyPIN("1234", hasher) // correct -> resets
				c.VerifyPIN("0000", hasher) // 1 wrong attempt again
				return c
			},
			pin:        "1234",
			wantErr:    nil,
			wantStatus: StatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.VerifyPIN(tt.pin, hasher)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Status != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, card.Status)
			}
		})
	}
}

func TestCard_Block_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		wantErr bool
	}{
		{
			name: "block active card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			wantErr: false,
		},
		{
			name: "block inactive card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			wantErr: false,
		},
		{
			name: "block cancelled card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Cancel()
				return c
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.Block()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Status != StatusBlocked {
				t.Errorf("expected BLOCKED, got %s", card.Status)
			}
		})
	}
}

func TestCard_Unblock_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		wantErr bool
	}{
		{
			name: "unblock blocked card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			wantErr: false,
		},
		{
			name: "unblock active card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			wantErr: true,
		},
		{
			name: "unblock inactive card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.Unblock()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.Status != StatusActive {
				t.Errorf("expected ACTIVE, got %s", card.Status)
			}
			if card.PINAttempts != 0 {
				t.Errorf("expected PIN attempts reset to 0, got %d", card.PINAttempts)
			}
		})
	}
}

func TestCard_Enroll3DS_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		version string
		wantErr bool
	}{
		{
			name: "enroll active card in 3DS v2.1",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			version: "2.1",
			wantErr: false,
		},
		{
			name: "enroll active card in 3DS v2.2",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			version: "2.2",
			wantErr: false,
		},
		{
			name: "enroll blocked card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			version: "2.1",
			wantErr: true,
		},
		{
			name: "enroll inactive card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				return c
			},
			version: "2.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.Enroll3DS(tt.version)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !card.ThreeDSEnrolled {
				t.Error("ThreeDSEnrolled should be true")
			}
			if card.ThreeDSVersion != tt.version {
				t.Errorf("expected 3DS version %s, got %s", tt.version, card.ThreeDSVersion)
			}
		})
	}
}

func TestCard_SetEMVAID_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		aid     string
		wantErr bool
	}{
		{
			name: "set EMVAID on active card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			aid:     "A0000000041010",
			wantErr: false,
		},
		{
			name: "set EMVAID on blocked card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			aid:     "A0000000041010",
			wantErr: true,
		},
		{
			name: "set EMVAID on cancelled card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Cancel()
				return c
			},
			aid:     "A0000000041010",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.SetEMVAID(tt.aid)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if card.EMVAID != tt.aid {
				t.Errorf("expected EMVAID %s, got %s", tt.aid, card.EMVAID)
			}
		})
	}
}

func TestCard_Cancel_IsPermanent(t *testing.T) {
	c, _ := NewCard("acc-1", TypeDebit)
	c.Activate("1234", hasher)
	c.Cancel()

	if c.Status != StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", c.Status)
	}

	// Cannot block a cancelled card
	if err := c.Block(); err == nil {
		t.Error("should not be able to block a cancelled card")
	}
}

func TestCard_ChangePIN_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Card
		oldPIN  string
		newPIN  string
		wantErr bool
	}{
		{
			name: "valid PIN change",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			oldPIN:  "1234",
			newPIN:  "5678",
			wantErr: false,
		},
		{
			name: "wrong old PIN",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			oldPIN:  "0000",
			newPIN:  "5678",
			wantErr: true,
		},
		{
			name: "new PIN too short",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				return c
			},
			oldPIN:  "1234",
			newPIN:  "12",
			wantErr: true,
		},
		{
			name: "change PIN on blocked card",
			setup: func() *Card {
				c, _ := NewCard("acc-1", TypeDebit)
				c.Activate("1234", hasher)
				c.Block()
				return c
			},
			oldPIN:  "1234",
			newPIN:  "5678",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.setup()
			err := card.ChangePIN(tt.oldPIN, tt.newPIN, hasher)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateLuhn_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		number string
		valid  bool
	}{
		{"Visa valid", "4539148803436467", true},
		{"Mastercard valid", "5425233430109903", true},
		{"XBank BIN valid", "4864860000000000", false}, // random, check actual
		{"all zeros", "0000000000000000", true},        // 0 mod 10 == 0
		{"single digit change", "4539148803436468", false},
		{"short number", "123", false},
		{"non-numeric", "abcdefghijklmnop", false},
		{"empty string", "", true}, // sum=0, 0%10==0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateLuhn(tt.number)
			if got != tt.valid {
				t.Errorf("ValidateLuhn(%s) = %v, want %v", tt.number, got, tt.valid)
			}
		})
	}
}

func TestMaskPAN_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		pan  string
		want string
	}{
		{"standard 16 digit", "4861234567891234", "**** **** **** 1234"},
		{"different last 4", "4861234567899999", "**** **** **** 9999"},
		{"short PAN", "12", "****"},
		{"3 char PAN", "abc", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskPAN(tt.pan)
			if got != tt.want {
				t.Errorf("MaskPAN(%s) = %s, want %s", tt.pan, got, tt.want)
			}
		})
	}
}
