package crypto

import (
	"testing"
	"time"
)

func BenchmarkAES_Encrypt(b *testing.B) {
	key, _ := GenerateHMACSecret() // 32-byte hex key
	enc, err := NewAESEncryptor(key)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := "4111111111111111" // typical PAN

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encrypt(plaintext)
	}
}

func BenchmarkAES_Decrypt(b *testing.B) {
	key, _ := GenerateHMACSecret()
	enc, _ := NewAESEncryptor(key)

	ciphertext, _ := enc.Encrypt("4111111111111111")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decrypt(ciphertext)
	}
}

func BenchmarkAES_RoundTrip(b *testing.B) {
	key, _ := GenerateHMACSecret()
	enc, _ := NewAESEncryptor(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct, _ := enc.Encrypt("4111111111111111")
		enc.Decrypt(ct)
	}
}

func BenchmarkHMAC_Sign(b *testing.B) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"from_account_id":"acc-1","to_account_id":"acc-2","amount":50000,"currency":"UZS"}`)
	ts := time.Now().Unix()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.Sign(ts, body)
	}
}

func BenchmarkHMAC_Verify(b *testing.B) {
	secret, _ := GenerateHMACSecret()
	signer, _ := NewHMACSigner(secret, 5*time.Minute)

	body := []byte(`{"from_account_id":"acc-1","to_account_id":"acc-2","amount":50000,"currency":"UZS"}`)
	ts := time.Now().Unix()
	sig := signer.Sign(ts, body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.Verify(sig, ts, body)
	}
}

func BenchmarkTokenizer_Tokenize(b *testing.B) {
	key, _ := GenerateHMACSecret()
	enc, _ := NewAESEncryptor(key)
	tok := NewTokenizer(enc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok.Tokenize("4111111111111111")
	}
}
