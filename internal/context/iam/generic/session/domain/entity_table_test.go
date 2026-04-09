package domain

import (
	"testing"
	"time"
)

func TestNewSession_TableDriven(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name             string
		userID           string
		refreshTokenHash string
		userAgent        string
		ipAddress        string
		expiresAt        time.Time
		wantErr          bool
		wantErrVal       error
	}{
		{
			name:             "valid session",
			userID:           "user-1",
			refreshTokenHash: "hashed_token_abc",
			userAgent:        "Mozilla/5.0",
			ipAddress:        "192.168.1.1",
			expiresAt:        future,
			wantErr:          false,
		},
		{
			name:             "valid session with empty user agent",
			userID:           "user-2",
			refreshTokenHash: "hashed_token_def",
			userAgent:        "",
			ipAddress:        "10.0.0.1",
			expiresAt:        future,
			wantErr:          false,
		},
		{
			name:             "valid session with empty IP",
			userID:           "user-3",
			refreshTokenHash: "hashed_token_ghi",
			userAgent:        "curl/7.68",
			ipAddress:        "",
			expiresAt:        future,
			wantErr:          false,
		},
		{
			name:             "empty user ID",
			userID:           "",
			refreshTokenHash: "hashed_token",
			userAgent:        "Mozilla/5.0",
			ipAddress:        "127.0.0.1",
			expiresAt:        future,
			wantErr:          true,
		},
		{
			name:             "empty refresh token",
			userID:           "user-1",
			refreshTokenHash: "",
			userAgent:        "Mozilla/5.0",
			ipAddress:        "127.0.0.1",
			expiresAt:        future,
			wantErr:          true,
			wantErrVal:       ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSession(tt.userID, tt.refreshTokenHash, tt.userAgent, tt.ipAddress, tt.expiresAt)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.wantErrVal != nil && err != tt.wantErrVal {
					t.Errorf("expected error %v, got %v", tt.wantErrVal, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.UserID != tt.userID {
				t.Errorf("expected userID %s, got %s", tt.userID, s.UserID)
			}
			if s.RefreshToken != tt.refreshTokenHash {
				t.Errorf("expected refreshToken %s, got %s", tt.refreshTokenHash, s.RefreshToken)
			}
			if s.UserAgent != tt.userAgent {
				t.Errorf("expected userAgent %s, got %s", tt.userAgent, s.UserAgent)
			}
			if s.IPAddress != tt.ipAddress {
				t.Errorf("expected ipAddress %s, got %s", tt.ipAddress, s.IPAddress)
			}
			if s.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
		})
	}
}

func TestSession_IsExpired_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired 1 hour ago",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired 1 second ago",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "expires in 1 hour",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expires in 30 days",
			expiresAt: time.Now().Add(30 * 24 * time.Hour),
			want:      false,
		},
		{
			name:      "expired long ago",
			expiresAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{ExpiresAt: tt.expiresAt}
			got := s.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_FieldsPersisted(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)
	s, err := NewSession("user-42", "token_hash_xyz", "Safari/15.0", "203.0.113.50", expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ID is not set by factory (set by repository/DB)
	if s.ID != "" {
		t.Error("ID should be empty (set by repository)")
	}

	if s.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}

	// ExpiresAt should match what we passed
	if !s.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expected ExpiresAt %v, got %v", expiresAt, s.ExpiresAt)
	}
}
