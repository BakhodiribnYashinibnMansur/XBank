// Package telegram provides a Telegram Bot API client for operational alerts.
// Protected by circuit breaker to prevent cascading failures.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/circuitbreaker"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Client sends messages via Telegram Bot API.
type Client struct {
	botToken   string
	chatID     string
	httpClient *http.Client
	breaker    *circuitbreaker.Breaker
}

// Config holds Telegram bot configuration.
type Config struct {
	BotToken string // TELEGRAM_BOT_TOKEN env
	ChatID   string // TELEGRAM_CHAT_ID env
}

// NewClient creates a Telegram client with circuit breaker protection.
func NewClient(cfg Config) *Client {
	return &Client{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		breaker: circuitbreaker.New("telegram", 3, 120*time.Second),
	}
}

// SendMessage sends a text message to the configured chat.
func (c *Client) SendMessage(ctx context.Context, text string) error {
	return c.breaker.Execute(func() error {
		return c.sendRequest(ctx, "sendMessage", map[string]interface{}{
			"chat_id":    c.chatID,
			"text":       text,
			"parse_mode": "HTML",
		})
	})
}

// SendAlert sends a formatted alert message with severity and details.
func (c *Client) SendAlert(ctx context.Context, severity, title, details string) error {
	emoji := severityEmoji(severity)
	text := fmt.Sprintf("%s <b>[%s] %s</b>\n\n%s\n\n<i>%s</i>",
		emoji, severity, title, details, time.Now().Format("2006-01-02 15:04:05"))

	return c.SendMessage(ctx, text)
}

func (c *Client) sendRequest(ctx context.Context, method string, payload map[string]interface{}) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", c.botToken, method)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("telegram send failed", zap.Error(err))
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("telegram API error", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("telegram API status: %d", resp.StatusCode)
	}

	return nil
}

// IsConfigured returns true if bot token and chat ID are set.
func (c *Client) IsConfigured() bool {
	return c.botToken != "" && c.chatID != ""
}

func severityEmoji(severity string) string {
	switch severity {
	case "CRITICAL":
		return "🔴"
	case "HIGH":
		return "🟠"
	case "MEDIUM":
		return "🟡"
	case "LOW":
		return "🟢"
	default:
		return "ℹ️"
	}
}
