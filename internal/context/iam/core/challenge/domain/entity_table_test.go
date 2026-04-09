package domain

import (
	"testing"
	"time"
)

func TestNewChallenge_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		method   Method
		action   string
		metadata string
		wantErr  bool
	}{
		{
			name:     "valid PASSWORD challenge",
			userID:   "user-1",
			method:   MethodPassword,
			action:   "transfer",
			metadata: `{"amount":50000}`,
			wantErr:  false,
		},
		{
			name:     "valid TOTP challenge",
			userID:   "user-2",
			method:   MethodTOTP,
			action:   "change_pin",
			metadata: "",
			wantErr:  false,
		},
		{
			name:     "empty metadata is valid",
			userID:   "user-3",
			method:   MethodPassword,
			action:   "delete_account",
			metadata: "",
			wantErr:  false,
		},
		{
			name:     "empty user ID",
			userID:   "",
			method:   MethodPassword,
			action:   "transfer",
			metadata: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := NewChallenge(tt.userID, tt.method, tt.action, tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ch.ID == "" {
				t.Error("ID should be generated")
			}
			if ch.UserID != tt.userID {
				t.Errorf("expected userID %s, got %s", tt.userID, ch.UserID)
			}
			if ch.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, ch.Method)
			}
			if ch.Action != tt.action {
				t.Errorf("expected action %s, got %s", tt.action, ch.Action)
			}
			if ch.Metadata != tt.metadata {
				t.Errorf("expected metadata %s, got %s", tt.metadata, ch.Metadata)
			}
			if ch.Status != StatusPending {
				t.Errorf("expected PENDING, got %s", ch.Status)
			}
			if ch.Token != "" {
				t.Error("token should be empty before verification")
			}
			if ch.VerifiedAt != nil {
				t.Error("VerifiedAt should be nil before verification")
			}
			if ch.ExpiresAt.Before(time.Now()) {
				t.Error("ExpiresAt should be in the future")
			}
		})
	}
}

func TestChallenge_Verify_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Challenge
		wantErr    bool
		wantStatus Status
	}{
		{
			name: "verify pending challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				return ch
			},
			wantErr:    false,
			wantStatus: StatusVerified,
		},
		{
			name: "verify already verified challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				return ch
			},
			wantErr:    true,
			wantStatus: StatusVerified,
		},
		{
			name: "verify expired challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.ExpiresAt = time.Now().Add(-1 * time.Minute)
				return ch
			},
			wantErr:    true,
			wantStatus: StatusExpired,
		},
		{
			name: "verify failed challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Fail()
				return ch
			},
			wantErr:    true,
			wantStatus: StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setup()
			err := ch.Verify()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if ch.Status != tt.wantStatus {
					t.Errorf("expected status %s, got %s", tt.wantStatus, ch.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ch.Status != StatusVerified {
				t.Errorf("expected VERIFIED, got %s", ch.Status)
			}
			if ch.Token == "" {
				t.Error("token should be generated after verification")
			}
			if ch.VerifiedAt == nil {
				t.Error("VerifiedAt should be set after verification")
			}
		})
	}
}

func TestChallenge_IsTokenValid_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*Challenge, string) // returns challenge and token to test
		want  bool
	}{
		{
			name: "valid token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				return ch, ch.Token
			},
			want: true,
		},
		{
			name: "wrong token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				return ch, "wrong-token"
			},
			want: false,
		},
		{
			name: "empty token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				return ch, ""
			},
			want: false,
		},
		{
			name: "expired token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				token := ch.Token
				ch.ExpiresAt = time.Now().Add(-1 * time.Second)
				return ch, token
			},
			want: false,
		},
		{
			name: "pending challenge token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				return ch, "any-token"
			},
			want: false,
		},
		{
			name: "failed challenge token",
			setup: func() (*Challenge, string) {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				token := ch.Token
				ch.Fail()
				return ch, token
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, token := tt.setup()
			got := ch.IsTokenValid(token)
			if got != tt.want {
				t.Errorf("IsTokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChallenge_Fail_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Challenge
		checkStatus Status
	}{
		{
			name: "fail pending challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				return ch
			},
			checkStatus: StatusFailed,
		},
		{
			name: "fail verified challenge",
			setup: func() *Challenge {
				ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
				ch.Verify()
				return ch
			},
			checkStatus: StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setup()
			ch.Fail()
			if ch.Status != tt.checkStatus {
				t.Errorf("expected status %s, got %s", tt.checkStatus, ch.Status)
			}
		})
	}
}

func TestChallenge_VerifyExtendsExpiry(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	originalExpiry := ch.ExpiresAt

	ch.Verify()

	// After verification, expiry should be extended (TokenTTL > DefaultTTL typically)
	if !ch.ExpiresAt.After(originalExpiry) {
		t.Error("verification should extend the expiry for token usage")
	}
}

func TestMethod_Constants(t *testing.T) {
	tests := []struct {
		method Method
		want   string
	}{
		{MethodPassword, "PASSWORD"},
		{MethodTOTP, "TOTP"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if string(tt.method) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.method)
			}
		})
	}
}

func TestStatus_Constants(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPending, "PENDING"},
		{StatusVerified, "VERIFIED"},
		{StatusExpired, "EXPIRED"},
		{StatusFailed, "FAILED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.status)
			}
		})
	}
}
