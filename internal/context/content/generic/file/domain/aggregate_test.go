package domain

import (
	"testing"
)

func TestNewFile_Success(t *testing.T) {
	f, err := NewFile(
		"abc123.jpg", "photo.jpg", "image/jpeg",
		1024000, "uploads/abc123.jpg", "https://cdn.xbank.uz/abc123.jpg", "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "abc123.jpg" {
		t.Errorf("Name expected abc123.jpg, got: %s", f.Name)
	}
	if f.OriginalName != "photo.jpg" {
		t.Errorf("OriginalName expected photo.jpg, got: %s", f.OriginalName)
	}
	if f.MimeType != "image/jpeg" {
		t.Errorf("MimeType expected image/jpeg, got: %s", f.MimeType)
	}
	if f.Size != 1024000 {
		t.Errorf("Size expected 1024000, got: %d", f.Size)
	}
	if f.UploadedBy != "user-1" {
		t.Errorf("UploadedBy expected user-1, got: %s", f.UploadedBy)
	}
	if f.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewFile_EmptyName(t *testing.T) {
	_, err := NewFile("", "photo.jpg", "image/jpeg", 1024, "path", "url", "user-1")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got: %v", err)
	}
}

func TestNewFile_EmptyOriginalName(t *testing.T) {
	_, err := NewFile("abc123.jpg", "", "image/jpeg", 1024, "path", "url", "user-1")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got: %v", err)
	}
}

func TestNewFile_SystemUpload(t *testing.T) {
	f, err := NewFile("sys-file.pdf", "report.pdf", "application/pdf", 5000, "path", "url", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.UploadedBy != "" {
		t.Errorf("system upload should have empty UploadedBy, got: %s", f.UploadedBy)
	}
}
