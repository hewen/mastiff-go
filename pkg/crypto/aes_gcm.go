// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// AESGCMCipher implements Cipher interface.
type AESGCMCipher struct {
	aead    cipher.AEAD
	Nonce   []byte
	AddData []byte
}

var aesNewCipher = aes.NewCipher

// NewAESGCMCipher creates a new AESGCMCipher.
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
		aead:  aead,
		Nonce: make([]byte, aead.NonceSize()),
	}, nil
}

// Encrypt encrypts data.
func (a *AESGCMCipher) Encrypt(data []byte) ([]byte, error) {
	return a.aead.Seal(nil, a.Nonce, data, a.AddData), nil
}

// Decrypt decrypts data.
func (a *AESGCMCipher) Decrypt(data []byte) ([]byte, error) {
	return a.aead.Open(nil, a.Nonce, data, a.AddData)
}

// CalcSize calculates the size of encrypted data.
func (a *AESGCMCipher) CalcSize(data []byte) int {
	return len(data) + a.aead.Overhead()
}
