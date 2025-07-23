package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESCipher_EncryptDecrypt(t *testing.T) {
	key := []byte("examplekey123456") // 16 bytes
	cipher, err := NewAESCipher(key)
	assert.NoError(t, err)

	cipher.Nonce = []byte("123456789012") // 12 bytes
	cipher.AddData = []byte("header")

	data := []byte("hello world")
	encrypted, err := cipher.Encrypt(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := cipher.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestAESCipher_InvalidKey(t *testing.T) {
	key := []byte("shortkey")
	_, err := NewAESCipher(key)
	assert.Error(t, err)
}

func TestAESCipher_CalcSize(t *testing.T) {
	key := []byte("examplekey123456") // 16 bytes
	cipher, err := NewAESCipher(key)
	assert.NoError(t, err)

	data := []byte("hello")
	expected := len(data) + cipher.aead.Overhead()
	got := cipher.CalcSize(data)
	assert.Equal(t, expected, got)
}
