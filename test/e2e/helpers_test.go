package e2e_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// apiResponse is the standard wrapper returned by XBank handlers.
type apiResponse struct {
	Data      json.RawMessage `json:"data"`
	Path      string          `json:"path"`
	Timestamp string          `json:"timestamp"`
}

// apiError is the standard error response.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// csrfToken caches the CSRF token obtained from a GET request.
var csrfToken string

// fetchCSRFToken does a GET request to obtain a CSRF token from the csrf_token cookie.
func fetchCSRFToken(t *testing.T, accessToken string) string {
	t.Helper()
	// Use a protected GET route that passes through CSRF middleware to get the token
	req := httptest.NewRequest("GET", "/api/v1/accounts/list", nil)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("fetching CSRF token: %v", err)
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
			return cookie.Value
		}
	}
	return csrfToken
}

// doRequest sends an HTTP request to the test Fiber app and returns the response.
func doRequest(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	// For mutating methods on non-exempt protected routes, fetch fresh CSRF token
	needsCSRF := (method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE") &&
		path != "/api/v1/auth/login" &&
		path != "/api/v1/auth/register" &&
		path != "/api/v1/auth/refresh" &&
		path != "/api/v1/auth/logout"
	if needsCSRF {
		fetchCSRFToken(t, token)
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshaling body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	}

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}

	// Update CSRF token from response cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
		}
	}

	rec := httptest.NewRecorder()
	rec.Code = resp.StatusCode
	for k, v := range resp.Header {
		for _, val := range v {
			rec.Header().Add(k, val)
		}
	}
	if resp.Body != nil {
		io.Copy(rec.Body, resp.Body)
		resp.Body.Close()
	}

	return rec
}

// doHMACRequest sends a request with HMAC signature headers.
func doHMACRequest(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshaling body: %v", err)
		}
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := signHMAC(timestamp, string(bodyBytes))

	// Ensure CSRF token
	if csrfToken == "" {
		fetchCSRFToken(t, token)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Signature-Timestamp", timestamp)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	}

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
		}
	}

	rec := httptest.NewRecorder()
	rec.Code = resp.StatusCode
	for k, v := range resp.Header {
		for _, val := range v {
			rec.Header().Add(k, val)
		}
	}
	if resp.Body != nil {
		io.Copy(rec.Body, resp.Body)
		resp.Body.Close()
	}

	return rec
}

// signHMAC computes HMAC-SHA256(secret, timestamp + "." + body).
func signHMAC(timestamp, body string) string {
	secretBytes, _ := hex.DecodeString(hmacSecret)
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(timestamp + "." + body))
	return hex.EncodeToString(mac.Sum(nil))
}

// parseResponse extracts the data from the standard API response wrapper.
func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()

	var resp apiResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parsing response wrapper: %v\nbody: %s", err, rec.Body.String())
	}

	if dest != nil && len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, dest); err != nil {
			t.Fatalf("parsing response data: %v\ndata: %s", err, string(resp.Data))
		}
	}
}

// parseError extracts error details from an error response.
func parseError(t *testing.T, rec *httptest.ResponseRecorder) apiError {
	t.Helper()
	var e apiError
	json.Unmarshal(rec.Body.Bytes(), &e)
	return e
}

// expectStatus asserts the response status code.
func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d\nbody: %s", rec.Code, want, rec.Body.String())
	}
}

// authTokens holds the result of a login.
type authTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
	} `json:"user"`
}

// registerAndLogin creates a user, logs in, and returns the tokens.
func registerAndLogin(t *testing.T, email, password, firstName string) authTokens {
	t.Helper()

	// Register
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   password,
		"first_name": firstName,
		"last_name":  "Test",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	// Login
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var tokens authTokens
	parseResponse(t, rec, &tokens)

	if tokens.AccessToken == "" {
		t.Fatal("expected access token after login")
	}
	return tokens
}
