package domain

import "testing"

func TestNewSystemError(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid", "DB_POOL_EXHAUSTED", false},
		{"empty code", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se, err := NewSystemError(tt.code, "connection pool exhausted", "CRITICAL", "SYSTEM")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if se.Code != tt.code {
				t.Errorf("code = %q, want %q", se.Code, tt.code)
			}
			if se.Resolution != StatusPending {
				t.Errorf("resolution = %q, want %q", se.Resolution, StatusPending)
			}
		})
	}
}

func TestSystemError_Resolve(t *testing.T) {
	se, _ := NewSystemError("TEST_ERR", "test", "LOW", "SYSTEM")

	if err := se.Resolve("admin@xbank.uz"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if se.Resolution != StatusResolved {
		t.Errorf("resolution = %q, want %q", se.Resolution, StatusResolved)
	}
	if se.ResolvedBy != "admin@xbank.uz" {
		t.Errorf("resolved_by = %q, want %q", se.ResolvedBy, "admin@xbank.uz")
	}
	if se.ResolvedAt == nil {
		t.Error("resolved_at should not be nil")
	}

	// double resolve should fail
	if err := se.Resolve("another"); err != ErrAlreadyResolved {
		t.Errorf("double resolve: got %v, want ErrAlreadyResolved", err)
	}
}

func TestSystemError_WithContext(t *testing.T) {
	se, _ := NewSystemError("TEST", "msg", "HIGH", "NETWORK")
	se.WithContext("req-123", "user-456", "192.168.1.1", "/api/v1/accounts", "POST", "stack...", map[string]string{"key": "val"})

	if se.RequestID != "req-123" {
		t.Errorf("request_id = %q", se.RequestID)
	}
	if se.Path != "/api/v1/accounts" {
		t.Errorf("path = %q", se.Path)
	}
	if se.Metadata["key"] != "val" {
		t.Errorf("metadata[key] = %q", se.Metadata["key"])
	}
}
