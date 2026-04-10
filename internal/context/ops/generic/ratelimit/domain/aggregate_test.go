package domain

import "testing"

func TestNewRateLimitRule(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		maxRequests   int
		windowSeconds int
		wantErr       bool
	}{
		{"valid", "/api/v1/auth/login", 10, 60, false},
		{"empty key", "", 10, 60, true},
		{"zero max requests", "/test", 0, 60, true},
		{"negative window", "/test", 10, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewRateLimitRule(tt.key, tt.maxRequests, tt.windowSeconds, "desc", true)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rule.Key != tt.key {
				t.Errorf("key = %q, want %q", rule.Key, tt.key)
			}
			if rule.MaxRequests != tt.maxRequests {
				t.Errorf("max_requests = %d, want %d", rule.MaxRequests, tt.maxRequests)
			}
			if !rule.Enabled {
				t.Error("should be enabled")
			}
		})
	}
}

func TestRateLimitRule_Update(t *testing.T) {
	rule, _ := NewRateLimitRule("/test", 10, 60, "old", true)

	rule.Update(100, 300, "updated", false)

	if rule.MaxRequests != 100 {
		t.Errorf("max_requests = %d, want 100", rule.MaxRequests)
	}
	if rule.WindowSeconds != 300 {
		t.Errorf("window_seconds = %d, want 300", rule.WindowSeconds)
	}
	if rule.Description != "updated" {
		t.Errorf("description = %q, want %q", rule.Description, "updated")
	}
	if rule.Enabled {
		t.Error("should be disabled after update")
	}
}
