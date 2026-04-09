package domain

import (
	"testing"
)

func TestNewAuditLog_Success(t *testing.T) {
	attrs := map[string]any{"key": "value"}
	log, err := NewAuditLog("User", "user-1", "LOGIN", "user-1", attrs, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.AggregateType != "User" {
		t.Errorf("AggregateType expected User, got: %s", log.AggregateType)
	}
	if log.AggregateID != "user-1" {
		t.Errorf("AggregateID expected user-1, got: %s", log.AggregateID)
	}
	if log.Action != "LOGIN" {
		t.Errorf("Action expected LOGIN, got: %s", log.Action)
	}
	if log.IPAddress != "192.168.1.1" {
		t.Errorf("IPAddress expected 192.168.1.1, got: %s", log.IPAddress)
	}
	if log.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewAuditLog_EmptyAggregateType(t *testing.T) {
	_, err := NewAuditLog("", "id-1", "CREATE", "actor-1", nil, "", "")
	if err != ErrInvalidAggregate {
		t.Errorf("expected ErrInvalidAggregate, got: %v", err)
	}
}

func TestNewAuditLog_EmptyAggregateID(t *testing.T) {
	_, err := NewAuditLog("User", "", "CREATE", "actor-1", nil, "", "")
	if err != ErrInvalidAggregate {
		t.Errorf("expected ErrInvalidAggregate, got: %v", err)
	}
}

func TestNewAuditLog_EmptyAction(t *testing.T) {
	_, err := NewAuditLog("User", "user-1", "", "actor-1", nil, "", "")
	if err != ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction, got: %v", err)
	}
}

func TestNewAuditLog_NilAttributes(t *testing.T) {
	log, err := NewAuditLog("User", "user-1", "DELETE", "actor-1", nil, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.Attributes == nil {
		t.Error("nil attributes should be initialized to empty map")
	}
}

func TestNewEndpointHistory_Success(t *testing.T) {
	h, err := NewEndpointHistory("GET", "/api/v1/accounts", 200, "user-1", "192.168.1.1", 45)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Method != "GET" {
		t.Errorf("Method expected GET, got: %s", h.Method)
	}
	if h.Path != "/api/v1/accounts" {
		t.Errorf("Path expected /api/v1/accounts, got: %s", h.Path)
	}
	if h.StatusCode != 200 {
		t.Errorf("StatusCode expected 200, got: %d", h.StatusCode)
	}
	if h.DurationMs != 45 {
		t.Errorf("DurationMs expected 45, got: %d", h.DurationMs)
	}
	if h.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}
