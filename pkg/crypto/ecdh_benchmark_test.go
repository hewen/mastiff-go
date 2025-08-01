package crypto

import (
	"testing"
)

func benchmarkECDHSharedSecret(b *testing.B, curveType ECDHCurveType) {
	cipherA, err := NewECDHCipher(curveType)
	if err != nil {
		b.Fatalf("failed to create cipher A: %v", err)
	}
	cipherB, err := NewECDHCipher(curveType)
	if err != nil {
		b.Fatalf("failed to create cipher B: %v", err)
	}

	pubB := cipherB.GetPublicKey()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cipherA.CalcSharedSecret(pubB)
		if err != nil {
			b.Fatalf("CalcSharedSecret failed: %v", err)
		}
	}
}

func BenchmarkECDHP256(b *testing.B) {
	benchmarkECDHSharedSecret(b, ECDHCurveP256)
}

func BenchmarkECDHP384(b *testing.B) {
	benchmarkECDHSharedSecret(b, ECDHCurveP384)
}

func BenchmarkECDHP521(b *testing.B) {
	benchmarkECDHSharedSecret(b, ECDHCurveP521)
}

func BenchmarkECDHX25519(b *testing.B) {
	benchmarkECDHSharedSecret(b, ECDHCurveX25519)
}
