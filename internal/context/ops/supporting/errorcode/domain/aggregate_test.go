package domain

import "testing"

func TestNewErrorCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid", "INSUFFICIENT_FUNDS", false},
		{"empty code", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ec, err := NewErrorCode(tt.code, "Insufficient funds", "Mablag' yetarli emas", "Недостаточно средств", "BUSINESS", "HIGH", 400, false, "Top up your account")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ec.Code != tt.code {
				t.Errorf("code = %q, want %q", ec.Code, tt.code)
			}
			if ec.HTTPStatus != 400 {
				t.Errorf("http_status = %d, want 400", ec.HTTPStatus)
			}
			if ec.Retryable {
				t.Error("retryable should be false")
			}
		})
	}
}

func TestErrorCode_Update(t *testing.T) {
	ec, _ := NewErrorCode("TEST", "en", "uz", "ru", "VALIDATION", "LOW", 400, false, "")

	newMsg := "updated message"
	newStatus := 422
	retryable := true
	ec.Update(&newMsg, nil, nil, nil, &newStatus, &retryable)

	if ec.MessageEn != "updated message" {
		t.Errorf("MessageEn = %q, want %q", ec.MessageEn, "updated message")
	}
	if ec.HTTPStatus != 422 {
		t.Errorf("HTTPStatus = %d, want 422", ec.HTTPStatus)
	}
	if !ec.Retryable {
		t.Error("Retryable should be true after update")
	}
	// nil fields should be unchanged
	if ec.MessageUz != "uz" {
		t.Errorf("MessageUz = %q, want %q (unchanged)", ec.MessageUz, "uz")
	}
}
