package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
)

func init() {
	logger.Init(true)
}

func TestGetSecret_Success(t *testing.T) {
	// Mock Vault server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/v1/secret/data/xbank/database" {
			if r.Header.Get("X-Vault-Token") != "test-token" {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"url": "postgres://user:pass@db:5432/xbank",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), Config{
		Address:   server.URL,
		Token:     "test-token",
		MountPath: "secret",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	secrets, err := client.GetSecret(context.Background(), "xbank/database")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	if secrets["url"] != "postgres://user:pass@db:5432/xbank" {
		t.Errorf("expected database url, got: %v", secrets)
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	_, err = client.GetSecret(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent secret")
	}
}

func TestGetSecret_InvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Header.Get("X-Vault-Token") != "valid-token" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []string{"permission denied"},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "wrong-token",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	_, err = client.GetSecret(context.Background(), "xbank/database")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestPutSecret_Success(t *testing.T) {
	var receivedData map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&receivedData)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.PutSecret(context.Background(), "xbank/test", map[string]string{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("PutSecret failed: %v", err)
	}

	data, ok := receivedData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("received data should have 'data' field")
	}
	if data["key"] != "value" {
		t.Errorf("expected key=value, got: %v", data)
	}
}

func TestHealthCheck_Sealed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err == nil {
		t.Error("expected error for sealed vault")
	}
}

func TestHealthCheck_NotInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}))
	defer server.Close()

	_, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err == nil {
		t.Error("expected error for uninitialized vault")
	}
}

func TestTransitEncrypt_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/v1/transit/encrypt/card-key" {
			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"ciphertext": "vault:v1:encrypted-data",
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ciphertext, err := client.TransitEncrypt(context.Background(), "card-key", "dGVzdA==")
	if err != nil {
		t.Fatalf("TransitEncrypt failed: %v", err)
	}

	if ciphertext != "vault:v1:encrypted-data" {
		t.Errorf("expected ciphertext, got: %s", ciphertext)
	}
}
