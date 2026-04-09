package domain

import (
	"testing"
)

func TestNewVerification_Success(t *testing.T) {
	v, err := NewVerification("user-1", DocPassport, "AB1234567", "John", "Doe", "1990-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", v.UserID)
	}
	if v.DocumentType != DocPassport {
		t.Errorf("DocumentType expected PASSPORT, got: %s", v.DocumentType)
	}
	if v.Status != StatusPending {
		t.Errorf("new verification should be PENDING, got: %s", v.Status)
	}
	if v.FirstName != "John" {
		t.Errorf("FirstName expected John, got: %s", v.FirstName)
	}
	if v.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewVerification_MissingFields(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		docNumber string
		firstName string
	}{
		{"missing user_id", "", "AB123", "John"},
		{"missing doc_number", "user-1", "", "John"},
		{"missing first_name", "user-1", "AB123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewVerification(tt.userID, DocPassport, tt.docNumber, tt.firstName, "Doe", "1990-01-15")
			if err == nil {
				t.Error("expected error for missing required field")
			}
		})
	}
}

func TestVerification_Approve(t *testing.T) {
	v, _ := NewVerification("user-1", DocIDCard, "ID123456", "Jane", "Smith", "1995-06-20")

	v.Approve("admin-1")

	if v.Status != StatusApproved {
		t.Errorf("Status expected APPROVED, got: %s", v.Status)
	}
	if v.ReviewedBy != "admin-1" {
		t.Errorf("ReviewedBy expected admin-1, got: %s", v.ReviewedBy)
	}
	if v.UpdatedAt.Before(v.CreatedAt) {
		t.Error("UpdatedAt should be >= CreatedAt after approval")
	}
}

func TestVerification_Reject(t *testing.T) {
	v, _ := NewVerification("user-1", DocDriverLicense, "DL999", "Bob", "Brown", "1985-03-10")

	v.Reject("admin-2", "document expired")

	if v.Status != StatusRejected {
		t.Errorf("Status expected REJECTED, got: %s", v.Status)
	}
	if v.ReviewedBy != "admin-2" {
		t.Errorf("ReviewedBy expected admin-2, got: %s", v.ReviewedBy)
	}
	if v.RejectedReason != "document expired" {
		t.Errorf("RejectedReason expected 'document expired', got: %s", v.RejectedReason)
	}
}

func TestDocumentTypes(t *testing.T) {
	tests := []struct {
		docType DocType
		want    string
	}{
		{DocPassport, "PASSPORT"},
		{DocIDCard, "ID_CARD"},
		{DocDriverLicense, "DRIVER_LICENSE"},
	}

	for _, tt := range tests {
		if string(tt.docType) != tt.want {
			t.Errorf("DocType expected %s, got: %s", tt.want, tt.docType)
		}
	}
}
