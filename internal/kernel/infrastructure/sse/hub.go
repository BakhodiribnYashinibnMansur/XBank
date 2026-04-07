package sse

import (
	"encoding/json"
	"sync"

	notification "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Hub - manages SSE connections per user
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]bool // userID → set of channels
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[chan []byte]bool),
	}
}

// Subscribe - register a client channel for a user
func (h *Hub) Subscribe(userID string) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan []byte, 64)
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan []byte]bool)
	}
	h.clients[userID][ch] = true

	logger.Log.Debug("SSE client subscribed", zap.String("user_id", userID))
	return ch
}

// Unsubscribe - remove a client channel
func (h *Hub) Unsubscribe(userID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if channels, ok := h.clients[userID]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(h.clients, userID)
		}
	}
}

// Send - send a notification to all connected clients of a user
func (h *Hub) Send(userID string, event notification.Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	channels, ok := h.clients[userID]
	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	for ch := range channels {
		select {
		case ch <- data:
		default:
			// Channel full, skip (client is slow)
			logger.Log.Warn("SSE client channel full, dropping event",
				zap.String("user_id", userID),
				zap.String("event_type", string(event.Type)),
			)
		}
	}
}

// ConnectedUsers - number of connected users (for metrics)
func (h *Hub) ConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
