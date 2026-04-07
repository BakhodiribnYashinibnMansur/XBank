package domain

import (
	"testing"
	"time"
)

func TestNewChallenge(t *testing.T) {
	ch, err := NewChallenge("user-1", MethodPassword, "transfer", `{"amount":50000}`)
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID == "" {
		t.Error("ID should be generated")
	}
	if ch.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", ch.UserID)
	}
	if ch.Status != StatusPending {
		t.Errorf("expected PENDING, got %s", ch.Status)
	}
	if ch.Method != MethodPassword {
		t.Errorf("expected PASSWORD, got %s", ch.Method)
	}
	if ch.ExpiresAt.Before(time.Now()) {
		t.Error("expires_at should be in the future")
	}
}

func TestNewChallenge_EmptyUserID(t *testing.T) {
	_, err := NewChallenge("", MethodPassword, "transfer", "")
	if err == nil {
		t.Error("should reject empty user_id")
	}
}

func TestChallenge_Verify(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	if err := ch.Verify(); err != nil {
		t.Fatal(err)
	}
	if ch.Status != StatusVerified {
		t.Errorf("expected VERIFIED, got %s", ch.Status)
	}
	if ch.Token == "" {
		t.Error("token should be generated after verify")
	}
	if ch.VerifiedAt == nil {
		t.Error("verified_at should be set")
	}
}

func TestChallenge_Verify_AlreadyVerified(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	ch.Verify()

	if err := ch.Verify(); err == nil {
		t.Error("should reject double verification")
	}
}

func TestChallenge_Verify_Expired(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	ch.ExpiresAt = time.Now().Add(-1 * time.Minute) // force expire

	if err := ch.Verify(); err == nil {
		t.Error("should reject expired challenge")
	}
	if ch.Status != StatusExpired {
		t.Errorf("expected EXPIRED, got %s", ch.Status)
	}
}

func TestChallenge_Fail(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	ch.Fail()
	if ch.Status != StatusFailed {
		t.Errorf("expected FAILED, got %s", ch.Status)
	}
}

func TestChallenge_IsTokenValid(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	ch.Verify()

	if !ch.IsTokenValid(ch.Token) {
		t.Error("valid token should pass")
	}
	if ch.IsTokenValid("wrong-token") {
		t.Error("wrong token should fail")
	}
}

func TestChallenge_IsTokenValid_Expired(t *testing.T) {
	ch, _ := NewChallenge("user-1", MethodPassword, "transfer", "")
	ch.Verify()
	ch.ExpiresAt = time.Now().Add(-1 * time.Second) // force expire

	if ch.IsTokenValid(ch.Token) {
		t.Error("expired token should fail")
	}
}
