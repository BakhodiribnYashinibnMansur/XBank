package card

import (
	"context"
	"testing"

	domainCard "github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mock"
)

func newTestService() *Service {
	return NewService(mock.NewCardRepository(), nil) // nil = no encryption in tests
}

func TestIssueCard(t *testing.T) {
	svc := newTestService()

	c, err := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID == "" {
		t.Error("ID should not be empty")
	}
	if c.Status != domainCard.StatusInactive {
		t.Errorf("expected INACTIVE, got: %s", c.Status)
	}
}

func TestActivate(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)

	activated, err := svc.Activate(context.Background(), c.ID, "1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activated.Status != domainCard.StatusActive {
		t.Errorf("expected ACTIVE, got: %s", activated.Status)
	}
}

func TestVerifyPIN_Success(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	svc.Activate(context.Background(), c.ID, "1234")

	if err := svc.VerifyPIN(context.Background(), c.ID, "1234"); err != nil {
		t.Errorf("correct PIN should pass: %v", err)
	}
}

func TestVerifyPIN_Wrong(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	svc.Activate(context.Background(), c.ID, "1234")

	if err := svc.VerifyPIN(context.Background(), c.ID, "0000"); err == nil {
		t.Error("wrong PIN should fail")
	}
}

func TestVerifyPIN_BruteForce_Blocks(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	svc.Activate(context.Background(), c.ID, "1234")

	svc.VerifyPIN(context.Background(), c.ID, "0000")
	svc.VerifyPIN(context.Background(), c.ID, "0000")
	svc.VerifyPIN(context.Background(), c.ID, "0000")

	// Card should be blocked now
	found, _ := svc.GetByID(context.Background(), c.ID)
	if found.Status != domainCard.StatusBlocked {
		t.Errorf("expected BLOCKED, got: %s", found.Status)
	}
}

func TestChangePIN(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	svc.Activate(context.Background(), c.ID, "1234")

	if err := svc.ChangePIN(context.Background(), c.ID, "1234", "5678"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// New PIN should work
	if err := svc.VerifyPIN(context.Background(), c.ID, "5678"); err != nil {
		t.Errorf("new PIN should work: %v", err)
	}
}

func TestBlock_Unblock(t *testing.T) {
	svc := newTestService()

	c, _ := svc.IssueCard(context.Background(), "acc-123", domainCard.TypeDebit)
	svc.Activate(context.Background(), c.ID, "1234")

	svc.Block(context.Background(), c.ID)
	found, _ := svc.GetByID(context.Background(), c.ID)
	if found.Status != domainCard.StatusBlocked {
		t.Error("should be BLOCKED")
	}

	svc.Unblock(context.Background(), c.ID)
	found, _ = svc.GetByID(context.Background(), c.ID)
	if found.Status != domainCard.StatusActive {
		t.Error("should be ACTIVE")
	}
}

func TestCardNotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetByID(context.Background(), "non-existent")
	if err != domainCard.ErrCardNotFound {
		t.Errorf("expected: %v, got: %v", domainCard.ErrCardNotFound, err)
	}
}
