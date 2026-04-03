package crypto

import (
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewAESEncryptor(key)
	tok := NewTokenizer(enc)

	pan := "4864861234567890"
	token, encrypted, lastFour, err := tok.Tokenize(pan)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(token, "tok_") {
		t.Errorf("token should start with tok_, got %s", token)
	}
	if len(token) != 24 {
		t.Errorf("token should be 24 chars, got %d", len(token))
	}
	if encrypted == "" || encrypted == pan {
		t.Error("encrypted PAN should differ from plaintext")
	}
	if lastFour != "7890" {
		t.Errorf("expected last four 7890, got %s", lastFour)
	}
}

func TestTokenize_UniqueTokens(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewAESEncryptor(key)
	tok := NewTokenizer(enc)

	t1, _, _, _ := tok.Tokenize("4864861234567890")
	t2, _, _, _ := tok.Tokenize("4864861234567890")

	if t1 == t2 {
		t.Error("same PAN should produce different tokens")
	}
}

func TestDetokenize(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewAESEncryptor(key)
	tok := NewTokenizer(enc)

	pan := "4864861234567890"
	_, encrypted, _, _ := tok.Tokenize(pan)

	decrypted, err := tok.Detokenize(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != pan {
		t.Errorf("expected %s, got %s", pan, decrypted)
	}
}

func TestTokenize_ShortPAN(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewAESEncryptor(key)
	tok := NewTokenizer(enc)

	_, _, _, err := tok.Tokenize("123")
	if err == nil {
		t.Error("should reject PAN shorter than 4 chars")
	}
}
