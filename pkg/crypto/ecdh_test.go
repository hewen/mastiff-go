package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECDHCipher_SharedSecret(t *testing.T) {
	alice, err := NewECDHCipher(ECDHCurveP256)
	assert.NoError(t, err)
	bob, err := NewECDHCipher(ECDHCurveP256)
	assert.NoError(t, err)

	alicePub := alice.GetPublicKey()
	bobPub := bob.GetPublicKey()

	secret1, err := alice.CalcSharedSecret(bobPub)
	assert.NoError(t, err)
	secret2, err := bob.CalcSharedSecret(alicePub)
	assert.NoError(t, err)

	assert.Equal(t, secret1, secret2)
}

func TestECDHCipher_CalcSharedSecret_InvalidKey(t *testing.T) {
	cipher, err := NewECDHCipher(ECDHCurveP256)
	assert.NoError(t, err)

	_, err = cipher.CalcSharedSecret(nil)
	assert.Error(t, err)

	_, err = cipher.CalcSharedSecret([]byte{0x00, 0x01, 0x02})
	assert.Error(t, err)
}

func TestNewECDHCipher_UnsupportedCurve(t *testing.T) {
	_, err := NewECDHCipher(ECDHCurveType(999))
	assert.Error(t, err)
}

func TestCurveFromType(t *testing.T) {
	curve, err := curveFromType(ECDHCurveP256)
	assert.NoError(t, err)
	assert.NotNil(t, curve)

	curve, err = curveFromType(ECDHCurveP384)
	assert.NoError(t, err)
	assert.NotNil(t, curve)

	curve, err = curveFromType(ECDHCurveP521)
	assert.NoError(t, err)
	assert.NotNil(t, curve)

	curve, err = curveFromType(ECDHCurveType(999))
	assert.Error(t, err)
	assert.Nil(t, curve)
}
