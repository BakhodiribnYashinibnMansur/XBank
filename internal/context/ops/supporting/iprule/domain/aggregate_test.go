package domain

import (
	"testing"
	"time"
)

func TestNewIPRule(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		ruleType  RuleType
		reason    string
		createdBy string
		wantErr   bool
	}{
		{"valid allow", "192.168.1.1", RuleTypeAllow, "office IP", "admin", false},
		{"valid deny", "10.0.0.1", RuleTypeDeny, "suspicious", "admin", false},
		{"empty IP", "", RuleTypeAllow, "test", "admin", true},
		{"invalid type", "1.2.3.4", RuleType("BLOCK"), "test", "admin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewIPRule(tt.ip, tt.ruleType, tt.reason, tt.createdBy, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rule.IPAddress != tt.ip {
				t.Errorf("ip = %q, want %q", rule.IPAddress, tt.ip)
			}
			if rule.RuleType != tt.ruleType {
				t.Errorf("type = %q, want %q", rule.RuleType, tt.ruleType)
			}
		})
	}
}

func TestIPRule_IsExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{"nil expiry", nil, false},
		{"past expiry", &past, true},
		{"future expiry", &future, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, _ := NewIPRule("1.2.3.4", RuleTypeAllow, "test", "admin", tt.expiresAt)
			if got := rule.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
