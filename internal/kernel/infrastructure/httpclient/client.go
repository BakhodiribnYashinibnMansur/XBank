// Package httpclient provides an instrumented HTTP client for external API calls.
// Automatically logs slow responses, errors, and samples successes.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Client wraps http.Client with automatic logging and metrics.
type Client struct {
	http          *http.Client
	slowThreshold time.Duration
	ssrfGuard     *SSRFGuard
}

// NewClient creates an instrumented HTTP client.
func NewClient(timeout, slowThreshold time.Duration) *Client {
	return &Client{
		http:          &http.Client{Timeout: timeout},
		slowThreshold: slowThreshold,
	}
}

// NewClientWithSSRF creates an instrumented HTTP client with SSRF protection.
func NewClientWithSSRF(timeout, slowThreshold time.Duration, guard *SSRFGuard) *Client {
	return &Client{
		http:          &http.Client{Timeout: timeout},
		slowThreshold: slowThreshold,
		ssrfGuard:     guard,
	}
}

// Do executes an HTTP request with logging.
// If SSRFGuard is configured, validates the URL before sending.
func (c *Client) Do(ctx context.Context, req *http.Request, operation string) (*http.Response, []byte, error) {
	if c.ssrfGuard != nil {
		if err := c.ssrfGuard.Validate(req.URL.String()); err != nil {
			logger.Log.Warn("SSRF blocked outbound request",
				zap.String("operation", operation),
				zap.String("url", req.URL.String()),
				zap.Error(err),
			)
			return nil, nil, fmt.Errorf("httpclient %s: %w", operation, err)
		}
	}

	start := time.Now()

	resp, err := c.http.Do(req.WithContext(ctx))
	elapsed := time.Since(start)

	if err != nil {
		logger.Log.Error("external API call failed",
			zap.String("operation", operation),
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Duration("duration", elapsed),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("httpclient %s: %w", operation, err)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	if elapsed > c.slowThreshold {
		logger.Log.Warn("slow external API call",
			zap.String("operation", operation),
			zap.String("url", req.URL.String()),
			zap.Int("status", resp.StatusCode),
			zap.Duration("duration", elapsed),
		)
	}

	if resp.StatusCode >= 400 {
		// Truncate error response body for logging
		logBody := string(body)
		if len(logBody) > 1024 {
			logBody = logBody[:1024] + "..."
		}
		logger.Log.Warn("external API error response",
			zap.String("operation", operation),
			zap.Int("status", resp.StatusCode),
			zap.String("body", logBody),
		)
	}

	return resp, body, nil
}

// PostJSON sends a JSON POST request.
func (c *Client) PostJSON(ctx context.Context, url, operation string, payload interface{}) (*http.Response, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(ctx, req, operation)
}

// GetJSON sends a GET request.
func (c *Client) GetJSON(ctx context.Context, url, operation string) (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient request: %w", err)
	}
	return c.Do(ctx, req, operation)
}
