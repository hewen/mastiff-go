package crypto

import (
	"crypto/cipher"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type badBlock struct{}

func (badBlock) BlockSize() int      { return 8 }
func (badBlock) Encrypt(_, _ []byte) {}
func (badBlock) Decrypt(_, _ []byte) {}

func TestNewAESCipher_GCMInitFail(t *testing.T) {
	_, err := cipher.NewGCM(badBlock{})
	require.Error(t, err)

	originalAESNewCipher := aesNewCipher
	defer func() { aesNewCipher = originalAESNewCipher }()

	aesNewCipher = func(_ []byte) (cipher.Block, error) {
		return badBlock{}, nil
	}

	_, err = NewAESCipher([]byte("0123456789abcdef"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create GCM mode")
}
