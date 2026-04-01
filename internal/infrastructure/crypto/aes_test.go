package crypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	enc, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatal(err)
	}

	pan := "4864861234567890"

	ciphertext, err := enc.Encrypt(pan)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	// Ciphertext should not contain original PAN
	if ciphertext == pan {
		t.Error("ciphertext should differ from plaintext")
	}

	// Decrypt should recover original
	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if decrypted != pan {
		t.Errorf("expected %s, got %s", pan, decrypted)
	}
}

func TestEncryptDifferentNonce(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewAESEncryptor(key)

	c1, _ := enc.Encrypt("same-data")
	c2, _ := enc.Encrypt("same-data")

	// Same plaintext, different ciphertext (random nonce)
	if c1 == c2 {
		t.Error("two encryptions of same data should produce different ciphertexts")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	enc1, _ := NewAESEncryptor(key1)
	enc2, _ := NewAESEncryptor(key2)

	ciphertext, _ := enc1.Encrypt("secret")

	_, err := enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("decryption with wrong key should fail")
	}
}
