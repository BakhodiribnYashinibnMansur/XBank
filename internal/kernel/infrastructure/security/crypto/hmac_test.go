package crypto

import (
	"testing"
	"time"
)

func TestHMACSignAndVerify(t *testing.T) {
	secret, err := GenerateHMACSecret()
	if err != nil {
		t.Fatal(err)
	}

	signer, err := NewHMACSigner(secret, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`{"amount":"1000","to_account":"ACC-123"}`)
	timestamp := time.Now().Unix()

	signature := signer.Sign(timestamp, body)

	// Valid signature should pass
	if err := signer.Verify(signature, timestamp, body); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
}

func TestHMACTamperedBody(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"amount":"1000"}`)
	timestamp := time.Now().Unix()
	signature := signer.Sign(timestamp, body)

	// Tampered body should fail
	tampered := []byte(`{"amount":"100000"}`)
	if err := signer.Verify(signature, timestamp, tampered); err == nil {
		t.Error("tampered body should fail verification")
	}
}

func TestHMACTamperedTimestamp(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"amount":"1000"}`)
	timestamp := time.Now().Unix()
	signature := signer.Sign(timestamp, body)

	// Different timestamp should fail (signature mismatch)
	if err := signer.Verify(signature, timestamp+1, body); err == nil {
		t.Error("different timestamp should fail verification")
	}
}

func TestHMACExpiredTimestamp(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"amount":"1000"}`)
	// 10 minutes ago — exceeds 5 min maxClockSkew
	timestamp := time.Now().Add(-10 * time.Minute).Unix()
	signature := signer.Sign(timestamp, body)

	if err := signer.Verify(signature, timestamp, body); err == nil {
		t.Error("expired timestamp should fail verification")
	}
}

func TestHMACFutureTimestamp(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"amount":"1000"}`)
	// 10 minutes in the future
	timestamp := time.Now().Add(10 * time.Minute).Unix()
	signature := signer.Sign(timestamp, body)

	if err := signer.Verify(signature, timestamp, body); err == nil {
		t.Error("future timestamp beyond skew should fail")
	}
}

func TestHMACWrongSecret(t *testing.T) {
	secret1, _ := GenerateHMACSecret()
	secret2, _ := GenerateHMACSecret()

	signer1, _ := NewHMACSigner(secret1, 5*time.Minute)
	signer2, _ := NewHMACSigner(secret2, 5*time.Minute)

	body := []byte(`{"amount":"1000"}`)
	timestamp := time.Now().Unix()
	signature := signer1.Sign(timestamp, body)

	// Different secret should fail
	if err := signer2.Verify(signature, timestamp, body); err == nil {
		t.Error("wrong secret should fail verification")
	}
}

func TestHMACEmptyBody(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	timestamp := time.Now().Unix()
	signature := signer.Sign(timestamp, []byte{})

	// Empty body should work fine
	if err := signer.Verify(signature, timestamp, []byte{}); err != nil {
		t.Fatalf("empty body should verify: %v", err)
	}
}

func TestHMACDeterministic(t *testing.T) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`same-data`)
	timestamp := time.Now().Unix()

	sig1 := signer.Sign(timestamp, body)
	sig2 := signer.Sign(timestamp, body)

	// Same input = same signature (unlike AES with random nonce)
	if sig1 != sig2 {
		t.Error("HMAC should be deterministic for same input")
	}
}

func TestHMACInvalidSecret(t *testing.T) {
	// Too short
	_, err := NewHMACSigner("aabbccdd", 5*time.Minute)
	if err == nil {
		t.Error("short secret should fail")
	}

	// Invalid hex
	_, err = NewHMACSigner("not-hex", 5*time.Minute)
	if err == nil {
		t.Error("invalid hex should fail")
	}
}
