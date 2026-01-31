package store

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hewen/mastiff-go/config/storeconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitRedis(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	_, err = InitRedis(storeconf.RedisConfig{
		Addr: s.Addr(),
	})
	assert.Nil(t, err)

}

func TestInitTLSRedis(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)
	cert, _ := tls.X509KeyPair(certPEM, keyPEM)

	s, err := miniredis.RunTLS(&tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	})
	assert.Nil(t, err)
	_, err = InitRedis(storeconf.RedisConfig{
		Addr: s.Addr(),
		TLSConfig: &storeconf.TLSConfig{
			Enabled:            true,
			VersionTLS:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
	})
	assert.Nil(t, err)
}

func generateTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour * 24),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&priv.PublicKey,
		priv,
	)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return
}
