package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateTempPrivateKeyFile(t *testing.T) string {
	priv, err := rsa.GenerateKey(RandomReader, 2048)
	assert.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(priv)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	tmpFile, err := os.CreateTemp("", "rsa_*.pem")
	assert.NoError(t, err)

	err = pem.Encode(tmpFile, pemBlock)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	return tmpFile.Name()
}

func TestRSA_EncryptDecrypt(t *testing.T) {
	priv, _ := rsa.GenerateKey(RandomReader, 2048)
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	r, err := NewRSAFromPEM(pemBlock)
	assert.NoError(t, err)

	message := []byte("test message")
	ciphertext, err := r.Encrypt(message)
	assert.NoError(t, err)

	plaintext, err := r.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, message, plaintext)
}

func TestRSA_LoadFromFile(t *testing.T) {
	path := generateTempPrivateKeyFile(t)
	defer func() { _ = os.Remove(path) }()

	r, err := NewRSAFromFile(path)
	assert.NoError(t, err)

	message := []byte("hello")
	ciphertext, err := r.Encrypt(message)
	assert.NoError(t, err)

	plaintext, err := r.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, message, plaintext)
}

func TestRSA_InvalidPEM(t *testing.T) {
	r, err := NewRSAFromPEM([]byte("invalid-pem"))
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestNewRSAFromFileFileNotExist(t *testing.T) {
	_, err := NewRSAFromFile("/path/does/not/exist.pem")
	assert.Error(t, err)
}

func TestNewRSAFromFileInvalidPEM(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid.pem")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.Write([]byte("not a pem file"))
	assert.NoError(t, err)
	defer func() { _ = tmpFile.Close() }()

	_, err = NewRSAFromFile(tmpFile.Name())
	assert.Error(t, err)
}
