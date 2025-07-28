package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func TestAESIGECipher_EncryptDecrypt(t *testing.T) {
	key := randomBytes(1)
	iv := randomBytes(1)
	_, err := NewAESIGECipher(key, iv)
	assert.NotNil(t, err)

	key = randomBytes(1)
	iv = randomBytes(32)
	_, err = NewAESIGECipher(key, iv)
	assert.NotNil(t, err)

	key = randomBytes(32) // 256-bit AES key
	iv = randomBytes(32)  // 2 blocks for IGE mode

	cipher, err := NewAESIGECipher(key, iv)
	assert.Nil(t, err)

	// Test with various plaintext lengths
	lengths := []int{16, 32, 64, 128, 256}
	for _, l := range lengths {
		plaintext := randomBytes(l)
		ciphertext, err := cipher.Encrypt(plaintext)
		assert.Nil(t, err)

		decrypted, err := cipher.Decrypt(ciphertext)
		assert.Nil(t, err)

		assert.Equal(t, plaintext, decrypted)
	}
}

func TestAESIGECipher_InvalidInput(t *testing.T) {
	key := randomBytes(32)
	iv := randomBytes(32)

	cipher, err := NewAESIGECipher(key, iv)
	assert.Nil(t, err)

	invalidData := randomBytes(30) // not multiple of block size
	_, err = cipher.Encrypt(invalidData)
	assert.NotNil(t, err)

	_, err = cipher.Decrypt(invalidData)
	assert.NotNil(t, err)
}
