package domain

import (
	"testing"
)

func TestNewIntegration_Success(t *testing.T) {
	i, err := NewIntegration("Stripe", "https://api.stripe.com", "sk_test_xxx", StatusActive, "https://xbank.uz/webhook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i.Name != "Stripe" {
		t.Errorf("Name expected Stripe, got: %s", i.Name)
	}
	if i.BaseURL != "https://api.stripe.com" {
		t.Errorf("BaseURL mismatch, got: %s", i.BaseURL)
	}
	if i.Status != StatusActive {
		t.Errorf("Status expected ACTIVE, got: %s", i.Status)
	}
	if i.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewIntegration_MissingName(t *testing.T) {
	_, err := NewIntegration("", "https://api.example.com", "key", StatusActive, "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestNewIntegration_MissingBaseURL(t *testing.T) {
	_, err := NewIntegration("Test", "", "key", StatusActive, "")
	if err == nil {
		t.Error("expected error for empty base_url")
	}
}

func TestNewIntegration_InvalidStatus(t *testing.T) {
	_, err := NewIntegration("Test", "https://api.example.com", "key", "INVALID", "")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestNewIntegration_ValidStatuses(t *testing.T) {
	statuses := []Status{StatusActive, StatusInactive, StatusSuspended}
	for _, s := range statuses {
		_, err := NewIntegration("Test", "https://api.example.com", "key", s, "")
		if err != nil {
			t.Errorf("status %s should be valid, got error: %v", s, err)
		}
	}
}

func TestIntegration_Update(t *testing.T) {
	i, _ := NewIntegration("Test", "https://old.api.com", "old-key", StatusActive, "")

	i.Update("https://new.api.com", "new-key", StatusInactive, "https://webhook.url")

	if i.BaseURL != "https://new.api.com" {
		t.Errorf("BaseURL expected https://new.api.com, got: %s", i.BaseURL)
	}
	if i.APIKey != "new-key" {
		t.Errorf("APIKey expected new-key, got: %s", i.APIKey)
	}
	if i.Status != StatusInactive {
		t.Errorf("Status expected INACTIVE, got: %s", i.Status)
	}
	if i.WebhookURL != "https://webhook.url" {
		t.Errorf("WebhookURL mismatch, got: %s", i.WebhookURL)
	}
}
