package command

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
)

// --- In-memory scheduled transfer repo ---

type memoryScheduledRepo struct {
	items map[string]*transfer.ScheduledTransfer
}

func newMemoryScheduledRepo() *memoryScheduledRepo {
	return &memoryScheduledRepo{items: make(map[string]*transfer.ScheduledTransfer)}
}

func (r *memoryScheduledRepo) Create(_ context.Context, st *transfer.ScheduledTransfer) error {
	r.items[st.ID] = st
	return nil
}

func (r *memoryScheduledRepo) GetByID(_ context.Context, id string) (*transfer.ScheduledTransfer, error) {
	st, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return st, nil
}

func (r *memoryScheduledRepo) ListByUserID(_ context.Context, userID string, limit, offset int) ([]*transfer.ScheduledTransfer, int64, error) {
	var result []*transfer.ScheduledTransfer
	for _, st := range r.items {
		if st.UserID == userID {
			result = append(result, st)
		}
	}
	total := int64(len(result))
	if offset < len(result) {
		end := offset + limit
		if end > len(result) {
			end = len(result)
		}
		result = result[offset:end]
	} else {
		result = nil
	}
	return result, total, nil
}

func (r *memoryScheduledRepo) Update(_ context.Context, st *transfer.ScheduledTransfer) error {
	r.items[st.ID] = st
	return nil
}

func (r *memoryScheduledRepo) FetchDue(_ context.Context, limit int) ([]*transfer.ScheduledTransfer, error) {
	var result []*transfer.ScheduledTransfer
	for _, st := range r.items {
		if st.Status == transfer.ScheduledPending && !time.Now().Before(st.ExecuteAt) {
			result = append(result, st)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func TestSchedule_Success(t *testing.T) {
	repo := newMemoryScheduledRepo()
	svc := NewScheduledService(repo, nil)

	executeAt := time.Now().Add(24 * time.Hour)
	st, err := svc.Schedule(context.Background(), "user-1", "acc-1", "acc-2", 5000, domain.Currency("UZS"), "test", executeAt)
	if err != nil {
		t.Fatal(err)
	}
	if st.Status != transfer.ScheduledPending {
		t.Errorf("expected PENDING, got %s", st.Status)
	}
}

func TestSchedule_PastDate(t *testing.T) {
	repo := newMemoryScheduledRepo()
	svc := NewScheduledService(repo, nil)

	executeAt := time.Now().Add(-1 * time.Hour)
	_, err := svc.Schedule(context.Background(), "user-1", "acc-1", "acc-2", 5000, domain.Currency("UZS"), "test", executeAt)
	if err == nil {
		t.Error("should reject past date")
	}
}

func TestSchedule_SameAccount(t *testing.T) {
	repo := newMemoryScheduledRepo()
	svc := NewScheduledService(repo, nil)

	executeAt := time.Now().Add(24 * time.Hour)
	_, err := svc.Schedule(context.Background(), "user-1", "acc-1", "acc-1", 5000, domain.Currency("UZS"), "test", executeAt)
	if err == nil {
		t.Error("should reject same account transfer")
	}
}

func TestCancel_Success(t *testing.T) {
	repo := newMemoryScheduledRepo()
	svc := NewScheduledService(repo, nil)

	executeAt := time.Now().Add(24 * time.Hour)
	st, _ := svc.Schedule(context.Background(), "user-1", "acc-1", "acc-2", 5000, domain.Currency("UZS"), "test", executeAt)

	err := svc.Cancel(context.Background(), st.ID, "user-1")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := svc.GetByID(context.Background(), st.ID)
	if updated.Status != transfer.ScheduledCancelled {
		t.Errorf("expected CANCELLED, got %s", updated.Status)
	}
}

func TestCancel_WrongUser(t *testing.T) {
	repo := newMemoryScheduledRepo()
	svc := NewScheduledService(repo, nil)

	executeAt := time.Now().Add(24 * time.Hour)
	st, _ := svc.Schedule(context.Background(), "user-1", "acc-1", "acc-2", 5000, domain.Currency("UZS"), "test", executeAt)

	err := svc.Cancel(context.Background(), st.ID, "user-2")
	if err == nil {
		t.Error("should reject cancel by wrong user")
	}
}
