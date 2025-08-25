// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// AESGCMCipher implements AES-GCM encryption/decryption with nonce management.
type AESGCMCipher struct {
	aead    cipher.AEAD
	AddData []byte
}

var aesNewCipher = aes.NewCipher

// NewAESGCMCipher creates a new AESGCMCipher with the given key.
func NewAESGCMCipher(key []byte) (*AESGCMCipher, error) {
	block, err := aesNewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	return &AESGCMCipher{
		aead: aead,
	}, nil
}

// Encrypt encrypts plaintext, prepending a random nonce to the ciphertext.
func (a *AESGCMCipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := a.aead.Seal(nil, nonce, plaintext, a.AddData)
	// Prepend nonce to ciphertext
	return append(nonce, ciphertext...), nil
}

// Decrypt decrypts ciphertext which must have nonce prepended.
func (a *AESGCMCipher) Decrypt(data []byte) ([]byte, error) {
	nonceSize := a.aead.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]
	plaintext, err := a.aead.Open(nil, nonce, ciphertext, a.AddData)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}
	return plaintext, nil
}

// CalcSize returns encrypted data length given plaintext length.
func (a *AESGCMCipher) CalcSize(plaintext []byte) int {
	return len(plaintext) + a.aead.Overhead() + a.aead.NonceSize()
}
