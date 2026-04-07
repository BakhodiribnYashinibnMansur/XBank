// Package main generates ECDSA P-256 key pairs for JWT ES256 signing.
//
// Usage:
//
//	go run ./cmd/keygen                    # default: keys/ directory
//	go run ./cmd/keygen -out-dir config/keys
//	go run ./cmd/keygen -force             # overwrite existing keys
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const defaultOutDir = "keys"

func main() {
	outDir := flag.String("out-dir", defaultOutDir, "directory to write key files into")
	force := flag.Bool("force", false, "overwrite existing PEM files")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		fatal(fmt.Errorf("create out-dir: %w", err))
	}

	privPath := filepath.Join(*outDir, "private.pem")
	pubPath := filepath.Join(*outDir, "public.pem")

	if !*force {
		if _, err := os.Stat(privPath); err == nil {
			fatal(fmt.Errorf("key already exists at %s (use -force to overwrite)", privPath))
		}
	}

	// Generate ECDSA P-256 key pair
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fatal(fmt.Errorf("generate key: %w", err))
	}

	// Write private key
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		fatal(fmt.Errorf("marshal private key: %w", err))
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		fatal(fmt.Errorf("write private key: %w", err))
	}
	fmt.Fprintf(os.Stderr, "wrote %s (mode 0600)\n", privPath)

	// Write public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		fatal(fmt.Errorf("marshal public key: %w", err))
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
		fatal(fmt.Errorf("write public key: %w", err))
	}
	fmt.Fprintf(os.Stderr, "wrote %s (mode 0644)\n", pubPath)

	fmt.Println("JWT ES256 key pair generated successfully")
}

func fatal(err error) {
	if errors.Is(err, os.ErrExist) {
		return
	}
	fmt.Fprintf(os.Stderr, "keygen: %v\n", err)
	os.Exit(1)
}
