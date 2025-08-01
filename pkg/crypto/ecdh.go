// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/ecdh"
	"errors"
)

// ECDHCipher implements Cipher interface.
type ECDHCipher struct {
	privKey *ecdh.PrivateKey
	curve   ecdh.Curve
	pubKey  []byte
}

// ECDHCurveType represents the type of ECDH curve, default to P256.
type ECDHCurveType int

const (
	// ECDHCurveP256 is the type of P256 curve.
	ECDHCurveP256 ECDHCurveType = iota
	// ECDHCurveP384 is the type of P384 curve.
	ECDHCurveP384
	// ECDHCurveP521 is the type of P521 curve.
	ECDHCurveP521
)

var generateKey = func(curve ecdh.Curve) (*ecdh.PrivateKey, error) {
	return curve.GenerateKey(RandomReader)
}

func curveFromType(tp ECDHCurveType) (ecdh.Curve, error) {
	switch tp {
	case ECDHCurveP256:
		return ecdh.P256(), nil
	case ECDHCurveP384:
		return ecdh.P384(), nil
	case ECDHCurveP521:
		return ecdh.P521(), nil
	default:
		return nil, errors.New("unsupported curve type")
	}
}

// NewECDHCipher creates a new ECDHCipher.
func NewECDHCipher(tp ECDHCurveType) (*ECDHCipher, error) {
	curve, err := curveFromType(tp)
	if err != nil {
		return nil, err
	}
	priv, err := generateKey(curve)
	if err != nil {
		return nil, err
	}

	pubBytes := priv.PublicKey().Bytes()

	return &ECDHCipher{
		privKey: priv,
		pubKey:  pubBytes,
		curve:   curve,
	}, nil
}

// GetPublicKey returns the public key.
func (c *ECDHCipher) GetPublicKey() []byte {
	return c.pubKey
}

// CalcSharedSecret calculates the shared secret.
func (c *ECDHCipher) CalcSharedSecret(peerPubKey []byte) ([]byte, error) {
	if len(peerPubKey) == 0 {
		return nil, ErrInvalidPublicKey
	}
	peerPub, err := c.curve.NewPublicKey(peerPubKey)
	if err != nil {
		return nil, err
	}

	secret, err := c.privKey.ECDH(peerPub)
	if err != nil {
		return nil, err
	}
	return secret, nil
}
