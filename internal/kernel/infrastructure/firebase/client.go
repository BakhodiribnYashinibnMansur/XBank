// Package firebase provides Firebase Cloud Messaging (FCM) integration
// for sending push notifications to mobile and web clients.
// Protected by circuit breaker.
package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/circuitbreaker"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Client sends push notifications via FCM HTTP v1 API.
type Client struct {
	serverKey  string
	httpClient *http.Client
	breaker    *circuitbreaker.Breaker
}

// Config holds FCM configuration.
type Config struct {
	ServerKey string // FCM server key from Firebase Console
}

// PushMessage represents a push notification payload.
type PushMessage struct {
	Token string            `json:"token"`            // Device FCM token
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
}

// NewClient creates an FCM client with circuit breaker.
func NewClient(cfg Config) *Client {
	return &Client{
		serverKey: cfg.ServerKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		breaker: circuitbreaker.New("firebase-fcm", 3, 60*time.Second),
	}
}

// Send sends a push notification to a single device.
func (c *Client) Send(ctx context.Context, msg PushMessage) error {
	return c.breaker.Execute(func() error {
		payload := map[string]interface{}{
			"to": msg.Token,
			"notification": map[string]string{
				"title": msg.Title,
				"body":  msg.Body,
			},
		}
		if len(msg.Data) > 0 {
			payload["data"] = msg.Data
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("fcm marshal: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			"https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("fcm request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "key="+c.serverKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			logger.Log.Error("fcm send failed", zap.Error(err))
			return fmt.Errorf("fcm send: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Log.Error("fcm API error", zap.Int("status", resp.StatusCode))
			return fmt.Errorf("fcm status: %d", resp.StatusCode)
		}
		return nil
	})
}

// SendToMany sends a push notification to multiple devices.
func (c *Client) SendToMany(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	for _, token := range tokens {
		if err := c.Send(ctx, PushMessage{
			Token: token,
			Title: title,
			Body:  body,
			Data:  data,
		}); err != nil {
			logger.Log.Warn("fcm send to device failed",
				zap.String("token_prefix", token[:min(8, len(token))]),
				zap.Error(err),
			)
		}
	}
	return nil
}

// IsConfigured returns true if server key is set.
func (c *Client) IsConfigured() bool {
	return c.serverKey != ""
}
