// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// AESCipher implements Cipher interface.
type AESCipher struct {
	aead    cipher.AEAD
	Nonce   []byte
	AddData []byte
}

var aesNewCipher = aes.NewCipher

// NewAESCipher creates a new AESCipher.
func NewAESCipher(key []byte) (*AESCipher, error) {
	block, err := aesNewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	return &AESCipher{
		aead:  aead,
		Nonce: make([]byte, aead.NonceSize()),
	}, nil
}

// Encrypt encrypts data.
func (a *AESCipher) Encrypt(data []byte) ([]byte, error) {
	return a.aead.Seal(nil, a.Nonce, data, a.AddData), nil
}

// Decrypt decrypts data.
func (a *AESCipher) Decrypt(data []byte) ([]byte, error) {
	return a.aead.Open(nil, a.Nonce, data, a.AddData)
}

// CalcSize calculates the size of encrypted data.
func (a *AESCipher) CalcSize(data []byte) int {
	return len(data) + a.aead.Overhead()
}
