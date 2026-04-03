package kafka

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	accountpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/accounts"
	transferpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/transfers"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	logger.Init(true)
}

func newTestMetadata(userID string) *commonpb.Metadata {
	return &commonpb.Metadata{
		EventId:   "evt-test",
		UserId:    userID,
		Timestamp: timestamppb.Now(),
		Source:    "test",
	}
}

// --- Account Handler Tests ---

func TestAccountEventHandler_HandleOpened(t *testing.T) {
	hub := sse.NewHub()
	ch := hub.Subscribe("user-1")
	defer hub.Unsubscribe("user-1", ch)

	handler := NewAccountEventHandler(hub)

	msg := &accountpb.AccountOpened{
		Metadata:      newTestMetadata("user-1"),
		AccountId:     "acc-1",
		UserId:        "user-1",
		AccountNumber: "1234567890",
		Currency:      "USD",
	}
	data, _ := proto.Marshal(msg)

	err := handler.Handle(context.Background(), "xbank.accounts.opened", []byte("acc-1"), data)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	select {
	case notification := <-ch:
		if len(notification) == 0 {
			t.Error("empty notification")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected SSE notification, got timeout")
	}
}

func TestAccountEventHandler_HandleCredited(t *testing.T) {
	hub := sse.NewHub()
	ch := hub.Subscribe("user-1")
	defer hub.Unsubscribe("user-1", ch)

	handler := NewAccountEventHandler(hub)

	msg := &accountpb.AccountCredited{
		Metadata:  newTestMetadata("user-1"),
		AccountId: "acc-1",
		Amount:    50000,
		Currency:  "UZS",
		Balance:   150000,
	}
	data, _ := proto.Marshal(msg)

	err := handler.Handle(context.Background(), "xbank.accounts.credited", []byte("acc-1"), data)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	select {
	case notification := <-ch:
		if len(notification) == 0 {
			t.Error("empty notification")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected SSE notification, got timeout")
	}
}

func TestAccountEventHandler_HandleClosed(t *testing.T) {
	hub := sse.NewHub()
	ch := hub.Subscribe("user-1")
	defer hub.Unsubscribe("user-1", ch)

	handler := NewAccountEventHandler(hub)

	msg := &accountpb.AccountClosed{
		Metadata:  newTestMetadata("user-1"),
		AccountId: "acc-1",
	}
	data, _ := proto.Marshal(msg)

	err := handler.Handle(context.Background(), "xbank.accounts.closed", []byte("acc-1"), data)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	select {
	case notification := <-ch:
		if len(notification) == 0 {
			t.Error("empty notification")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected SSE notification, got timeout")
	}
}

// --- Transfer Handler Tests ---

func TestTransferEventHandler_HandleCompleted(t *testing.T) {
	hub := sse.NewHub()
	ch := hub.Subscribe("user-1")
	defer hub.Unsubscribe("user-1", ch)

	handler := NewTransferEventHandler(hub)

	msg := &transferpb.TransferCompleted{
		Metadata:      newTestMetadata("user-1"),
		TransferId:    "txn-1",
		FromAccountId: "acc-1",
		ToAccountId:   "acc-2",
		Amount:        50000,
		Currency:      "UZS",
	}
	data, _ := proto.Marshal(msg)

	err := handler.Handle(context.Background(), "xbank.transfers.completed", []byte("txn-1"), data)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	select {
	case notification := <-ch:
		if len(notification) == 0 {
			t.Error("empty notification")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected SSE notification, got timeout")
	}
}

func TestTransferEventHandler_HandleFailed(t *testing.T) {
	hub := sse.NewHub()
	ch := hub.Subscribe("user-1")
	defer hub.Unsubscribe("user-1", ch)

	handler := NewTransferEventHandler(hub)

	msg := &transferpb.TransferFailed{
		Metadata:      newTestMetadata("user-1"),
		TransferId:    "txn-2",
		FromAccountId: "acc-1",
		ToAccountId:   "acc-2",
		Amount:        100000,
		Currency:      "UZS",
		Reason:        "insufficient funds",
	}
	data, _ := proto.Marshal(msg)

	err := handler.Handle(context.Background(), "xbank.transfers.failed", []byte("txn-2"), data)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	select {
	case notification := <-ch:
		if len(notification) == 0 {
			t.Error("empty notification")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected SSE notification, got timeout")
	}
}

// --- Error Cases ---

func TestAccountEventHandler_InvalidProto(t *testing.T) {
	hub := sse.NewHub()
	handler := NewAccountEventHandler(hub)

	err := handler.Handle(context.Background(), "xbank.accounts.opened", []byte("acc-1"), []byte("invalid-protobuf"))
	if err == nil {
		t.Error("expected error for invalid protobuf data")
	}
}

func TestTransferEventHandler_InvalidProto(t *testing.T) {
	hub := sse.NewHub()
	handler := NewTransferEventHandler(hub)

	err := handler.Handle(context.Background(), "xbank.transfers.completed", []byte("txn-1"), []byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid protobuf data")
	}
}

func TestAccountEventHandler_UnknownTopic(t *testing.T) {
	hub := sse.NewHub()
	handler := NewAccountEventHandler(hub)

	err := handler.Handle(context.Background(), "xbank.accounts.unknown", []byte("acc-1"), []byte{})
	if err != nil {
		t.Errorf("unknown topic should return nil, got: %v", err)
	}
}

// --- Utility Tests ---

func TestMatchSuffix(t *testing.T) {
	tests := []struct {
		topic  string
		suffix string
		expect bool
	}{
		{"xbank.accounts.opened", "opened", true},
		{"xbank.transfers.completed", "completed", true},
		{"xbank.accounts.opened", "closed", false},
		{"opened", "opened", true},
		{"xbank.accounts.frozen", "frozen", true},
	}

	for _, tc := range tests {
		got := matchSuffix(tc.topic, tc.suffix)
		if got != tc.expect {
			t.Errorf("matchSuffix(%q, %q) = %v, want %v", tc.topic, tc.suffix, got, tc.expect)
		}
	}
}

func TestConsumer_Subscribe(t *testing.T) {
	consumer := NewConsumer([]string{"localhost:9092"}, "test-group")

	var mu sync.Mutex
	calls := 0
	mockHandler := &mockMessageHandler{
		handleFn: func(ctx context.Context, topic string, key []byte, value []byte) error {
			mu.Lock()
			calls++
			mu.Unlock()
			return nil
		},
	}

	consumer.Subscribe("test.topic", mockHandler)

	if _, ok := consumer.handlers["test.topic"]; !ok {
		t.Error("handler should be registered for test.topic")
	}
}

type mockMessageHandler struct {
	handleFn func(ctx context.Context, topic string, key []byte, value []byte) error
}

func (m *mockMessageHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	return m.handleFn(ctx, topic, key, value)
}
