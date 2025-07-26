package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// AESIGECipher implements AES encryption using IGE mode.
type AESIGECipher struct {
	block cipher.Block
	iv    []byte // Must be 32 bytes: IV1 || IV2
}

// NewAESIGECipher creates a new AES cipher in IGE mode.
func NewAESIGECipher(key, iv []byte) (*AESIGECipher, error) {
	if len(iv) != aes.BlockSize*2 {
		return nil, errors.New("IGE IV must be 32 bytes (IV1 || IV2)")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESIGECipher{block: block, iv: append([]byte(nil), iv...)}, nil
}

// Encrypt encrypts data using AES-IGE mode.
// Input length must be a multiple of 16 bytes.
func (c *AESIGECipher) Encrypt(src []byte) ([]byte, error) {
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	dst := make([]byte, len(src))
	iv1 := append([]byte(nil), c.iv[:aes.BlockSize]...)
	iv2 := append([]byte(nil), c.iv[aes.BlockSize:]...)

	for i := 0; i < len(src); i += aes.BlockSize {
		block := xor(src[i:i+aes.BlockSize], iv1)
		c.block.Encrypt(block, block)
		block = xor(block, iv2)
		copy(dst[i:], block)

		iv1, iv2 = block, src[i:i+aes.BlockSize]
	}

	return dst, nil
}

// Decrypt decrypts AES-IGE encrypted data.
func (c *AESIGECipher) Decrypt(src []byte) ([]byte, error) {
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("input not full blocks")
	}

	dst := make([]byte, len(src))
	iv1 := append([]byte(nil), c.iv[:aes.BlockSize]...)
	iv2 := append([]byte(nil), c.iv[aes.BlockSize:]...)

	for i := 0; i < len(src); i += aes.BlockSize {
		tmp := xor(src[i:i+aes.BlockSize], iv2)
		c.block.Decrypt(tmp, tmp)
		block := xor(tmp, iv1)
		copy(dst[i:], block)

		iv1, iv2 = src[i:i+aes.BlockSize], block
	}

	return dst, nil
}

func xor(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}
