package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Client wraps the HashiCorp Vault HTTP API for secret management.
// Supports KV v2 secrets engine for reading/writing key-value secrets.
type Client struct {
	addr       string
	token      string
	httpClient *http.Client
	mountPath  string // KV v2 mount path (default: "secret")
}

// Config holds Vault connection parameters.
type Config struct {
	Address   string // e.g. "http://vault:8200"
	Token     string // root or app-role token
	MountPath string // KV v2 mount (default: "secret")
}

// NewClient creates a Vault client and verifies connectivity.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.MountPath == "" {
		cfg.MountPath = "secret"
	}

	c := &Client{
		addr:      strings.TrimRight(cfg.Address, "/"),
		token:     cfg.Token,
		mountPath: cfg.MountPath,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Verify Vault is reachable and unsealed
	if err := c.healthCheck(ctx); err != nil {
		return nil, fmt.Errorf("vault health check failed: %w", err)
	}

	logger.Log.Info("Vault connected", zap.String("addr", cfg.Address))
	return c, nil
}

// GetSecret reads a secret from KV v2 at the given path.
// Returns the data map from the latest version.
func (c *Client) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.addr, c.mountPath, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("vault request build: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("vault secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("vault response decode: %w", err)
	}

	// Convert all values to strings
	secrets := make(map[string]string, len(result.Data.Data))
	for k, v := range result.Data.Data {
		secrets[k] = fmt.Sprintf("%v", v)
	}

	return secrets, nil
}

// PutSecret writes a secret to KV v2 at the given path.
func (c *Client) PutSecret(ctx context.Context, path string, data map[string]string) error {
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.addr, c.mountPath, path)

	payload := map[string]interface{}{
		"data": data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("vault payload marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("vault request build: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault put error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// TransitEncrypt encrypts plaintext using Vault's Transit engine.
// keyName is the named encryption key in Transit.
func (c *Client) TransitEncrypt(ctx context.Context, keyName string, plaintext string) (string, error) {
	url := fmt.Sprintf("%s/v1/transit/encrypt/%s", c.addr, keyName)

	payload := fmt.Sprintf(`{"plaintext":"%s"}`, plaintext)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("transit encrypt request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("transit encrypt: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Ciphertext string `json:"ciphertext"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("transit encrypt decode: %w", err)
	}

	return result.Data.Ciphertext, nil
}

// TransitDecrypt decrypts ciphertext using Vault's Transit engine.
func (c *Client) TransitDecrypt(ctx context.Context, keyName string, ciphertext string) (string, error) {
	url := fmt.Sprintf("%s/v1/transit/decrypt/%s", c.addr, keyName)

	payload := fmt.Sprintf(`{"ciphertext":"%s"}`, ciphertext)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("transit decrypt request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("transit decrypt: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Plaintext string `json:"plaintext"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("transit decrypt decode: %w", err)
	}

	return result.Data.Plaintext, nil
}

// healthCheck verifies Vault is reachable and initialized.
func (c *Client) healthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/sys/health", c.addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Vault returns 200 (initialized+unsealed), 429 (standby), 472 (DR secondary),
	// 473 (perf standby), 501 (not initialized), 503 (sealed)
	switch resp.StatusCode {
	case http.StatusOK, http.StatusTooManyRequests:
		return nil // healthy or standby
	case http.StatusNotImplemented:
		return fmt.Errorf("vault not initialized")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("vault is sealed")
	default:
		return fmt.Errorf("vault health status: %d", resp.StatusCode)
	}
}

// Close is a no-op for the HTTP client (no persistent connection to close).
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
