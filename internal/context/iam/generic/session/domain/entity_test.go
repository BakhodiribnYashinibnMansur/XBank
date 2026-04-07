package domain

import (
	"testing"
	"time"
)

func TestNewSession_Success(t *testing.T) {
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 kun
	s, err := NewSession("user-123", "hashed_token", "Mozilla/5.0", "127.0.0.1", expiresAt)
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if s.UserID != "user-123" {
		t.Errorf("UserID kutilgan: user-123, kelgan: %s", s.UserID)
	}
}

func TestNewSession_EmptyUserID(t *testing.T) {
	_, err := NewSession("", "hashed_token", "Mozilla/5.0", "127.0.0.1", time.Now())
	if err == nil {
		t.Error("Xatolik kutilgan edi")
	}
}

func TestNewSession_EmptyToken(t *testing.T) {
	_, err := NewSession("user-123", "", "Mozilla/5.0", "127.0.0.1", time.Now())
	if err != ErrInvalidToken {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrInvalidToken, err)
	}
}

func TestSession_IsExpired(t *testing.T) {
	// Muddati o'tgan session
	expired := &Session{ExpiresAt: time.Now().Add(-1 * time.Hour)}
	if !expired.IsExpired() {
		t.Error("Session expired bo'lishi kerak edi")
	}

	// Hali amal qilayotgan session
	valid := &Session{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if valid.IsExpired() {
		t.Error("Session hali expired bo'lmasligi kerak")
	}
}
