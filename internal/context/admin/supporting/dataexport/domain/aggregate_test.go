package domain

import (
	"testing"
)

func TestNewDataExport_Success(t *testing.T) {
	e, err := NewDataExport("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", e.UserID)
	}
	if e.Status != ExportPending {
		t.Errorf("Status expected PENDING, got: %s", e.Status)
	}
}

func TestNewDataExport_EmptyUserID(t *testing.T) {
	_, err := NewDataExport("")
	if err != ErrEmptyUserID {
		t.Errorf("expected ErrEmptyUserID, got: %v", err)
	}
}

func TestDataExport_StartProcessing(t *testing.T) {
	e, _ := NewDataExport("user-1")
	if err := e.StartProcessing(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != ExportProcessing {
		t.Errorf("Status expected PROCESSING, got: %s", e.Status)
	}
}

func TestDataExport_StartProcessing_InvalidState(t *testing.T) {
	e, _ := NewDataExport("user-1")
	e.StartProcessing()

	// Trying to start processing again should fail
	err := e.StartProcessing()
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestDataExport_Complete(t *testing.T) {
	e, _ := NewDataExport("user-1")
	e.StartProcessing()

	if err := e.Complete("https://storage.xbank.uz/exports/user-1.zip"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != ExportStatusCompleted {
		t.Errorf("Status expected COMPLETED, got: %s", e.Status)
	}
	if e.FileURL != "https://storage.xbank.uz/exports/user-1.zip" {
		t.Errorf("FileURL mismatch, got: %s", e.FileURL)
	}
}

func TestDataExport_Complete_InvalidState(t *testing.T) {
	e, _ := NewDataExport("user-1")

	// Cannot complete from PENDING (must be PROCESSING)
	err := e.Complete("url")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestDataExport_Fail(t *testing.T) {
	e, _ := NewDataExport("user-1")
	e.StartProcessing()

	if err := e.Fail("disk full"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != ExportFailed {
		t.Errorf("Status expected FAILED, got: %s", e.Status)
	}
	if e.ErrorMsg != "disk full" {
		t.Errorf("ErrorMsg expected 'disk full', got: %s", e.ErrorMsg)
	}
}

func TestDataExport_Fail_InvalidState(t *testing.T) {
	e, _ := NewDataExport("user-1")

	// Cannot fail from PENDING (must be PROCESSING)
	err := e.Fail("reason")
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestDataExport_StateMachine(t *testing.T) {
	// Test the complete happy path: PENDING -> PROCESSING -> COMPLETED
	e, _ := NewDataExport("user-1")
	if e.Status != ExportPending {
		t.Fatal("initial state should be PENDING")
	}

	e.StartProcessing()
	if e.Status != ExportProcessing {
		t.Fatal("state should be PROCESSING")
	}

	e.Complete("url")
	if e.Status != ExportStatusCompleted {
		t.Fatal("state should be COMPLETED")
	}
}
