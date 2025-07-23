// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// RSA implements Cipher interface.
type RSA struct {
	PrivateKey *rsa.PrivateKey
}

// NewRSAFromPEM creates a new RSA from PEM bytes.
func NewRSAFromPEM(pemBytes []byte) (*RSA, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return &RSA{PrivateKey: privKey}, nil
}

// NewRSAFromFile creates a new RSA from a file.
func NewRSAFromFile(path string) (*RSA, error) {
	file, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return NewRSAFromPEM(content)
}

// Encrypt encrypts data using PKCS#1 v1.5 padding.
func (r *RSA) Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, &r.PrivateKey.PublicKey, data)
}

// Decrypt decrypts data using PKCS#1 v1.5 padding.
func (r *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, r.PrivateKey, ciphertext)
}
