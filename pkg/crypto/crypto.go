// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/rand"
	"errors"
)

// RandomReader is a global variable for random number generation.
var RandomReader = rand.Reader

var (
	// ErrInvalidPublicKey is returned when the public key is invalid.
	ErrInvalidPublicKey = errors.New("invalid public key")
)

// Cipher is an interface for encrypting and decrypting data.
type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}
