package domain

import (
	"testing"
)

func TestNewNotification_Success(t *testing.T) {
	data := map[string]string{"transfer_id": "tx-1"}
	n, err := NewNotification("user-1", "Transfer Complete", "Your transfer was successful", NotificationInfo, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", n.UserID)
	}
	if n.Title != "Transfer Complete" {
		t.Errorf("Title mismatch, got: %s", n.Title)
	}
	if n.Type != NotificationInfo {
		t.Errorf("Type expected INFO, got: %s", n.Type)
	}
	if n.IsRead() {
		t.Error("new notification should be unread")
	}
}

func TestNewNotification_EmptyUserID(t *testing.T) {
	_, err := NewNotification("", "Title", "Message", NotificationInfo, nil)
	if err != ErrEmptyUserID {
		t.Errorf("expected ErrEmptyUserID, got: %v", err)
	}
}

func TestNewNotification_EmptyTitle(t *testing.T) {
	_, err := NewNotification("user-1", "", "Message", NotificationInfo, nil)
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got: %v", err)
	}
}

func TestNewNotification_EmptyMessage(t *testing.T) {
	_, err := NewNotification("user-1", "Title", "", NotificationInfo, nil)
	if err != ErrEmptyMessage {
		t.Errorf("expected ErrEmptyMessage, got: %v", err)
	}
}

func TestNewNotification_DefaultType(t *testing.T) {
	n, err := NewNotification("user-1", "Title", "Message", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Type != NotificationInfo {
		t.Errorf("empty type should default to INFO, got: %s", n.Type)
	}
}

func TestNotification_MarkAsRead(t *testing.T) {
	n, _ := NewNotification("user-1", "Title", "Message", NotificationInfo, nil)

	if n.IsRead() {
		t.Error("should be unread initially")
	}

	n.MarkAsRead()

	if !n.IsRead() {
		t.Error("should be read after MarkAsRead()")
	}
	if n.ReadAt == nil {
		t.Error("ReadAt should be set")
	}
}

func TestNotification_MarkAsRead_Idempotent(t *testing.T) {
	n, _ := NewNotification("user-1", "Title", "Message", NotificationInfo, nil)

	n.MarkAsRead()
	firstReadAt := n.ReadAt

	n.MarkAsRead() // second call should be no-op

	if n.ReadAt != firstReadAt {
		t.Error("MarkAsRead should be idempotent; ReadAt should not change on second call")
	}
}

func TestNotificationTypes(t *testing.T) {
	tests := []struct {
		nt   NotificationType
		want string
	}{
		{NotificationInfo, "INFO"},
		{NotificationWarning, "WARNING"},
		{NotificationAlert, "ALERT"},
	}

	for _, tt := range tests {
		if string(tt.nt) != tt.want {
			t.Errorf("NotificationType expected %s, got: %s", tt.want, tt.nt)
		}
	}
}
