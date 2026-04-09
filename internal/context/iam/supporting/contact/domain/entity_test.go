package domain

import (
	"testing"
)

func TestNewContact_Success(t *testing.T) {
	c, err := NewContact("owner-1", "contact-1", "My Friend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.OwnerID != "owner-1" {
		t.Errorf("OwnerID expected owner-1, got: %s", c.OwnerID)
	}
	if c.ContactID != "contact-1" {
		t.Errorf("ContactID expected contact-1, got: %s", c.ContactID)
	}
	if c.CustomName != "My Friend" {
		t.Errorf("CustomName expected My Friend, got: %s", c.CustomName)
	}
	if c.IsBlocked {
		t.Error("new contact should not be blocked")
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewContact_EmptyOwnerID(t *testing.T) {
	_, err := NewContact("", "contact-1", "Name")
	if err == nil {
		t.Error("expected error for empty owner_id")
	}
}

func TestNewContact_EmptyContactID(t *testing.T) {
	_, err := NewContact("owner-1", "", "Name")
	if err == nil {
		t.Error("expected error for empty contact_id")
	}
}

func TestNewContact_SelfContact(t *testing.T) {
	_, err := NewContact("user-1", "user-1", "Myself")
	if err != ErrContactSelf {
		t.Errorf("expected ErrContactSelf, got: %v", err)
	}
}

func TestNewContact_EmptyCustomName(t *testing.T) {
	c, err := NewContact("owner-1", "contact-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CustomName != "" {
		t.Errorf("CustomName expected empty, got: %s", c.CustomName)
	}
}
