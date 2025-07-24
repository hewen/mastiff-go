package compress

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkNoCompress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeNoCompress)
	}
}

func BenchmarkNoDecompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeNoCompress)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeNoCompress)
	}
}

func BenchmarkZlibCompress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeZlib)
	}
}

func BenchmarkZlibDecompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeZlib)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeZlib)
	}
}

func BenchmarkSnappyCompress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeSnappy)
	}
}

func BenchmarkSnappyDecompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeSnappy)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeSnappy)
	}
}

func BenchmarkLz4Compress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeLz4)
	}
}

func BenchmarkLz4Decompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeLz4)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeLz4)
	}
}

func BenchmarkZstdCompress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeZstd)
	}
}

func BenchmarkZstdDecompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeZstd)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeZstd)
	}
}

func BenchmarkBrotliCompress(b *testing.B) {
	data := getTestData(b)
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeBrotli)
	}
}

func BenchmarkBrotliDecompress(b *testing.B) {
	data := getTestData(b)
	compressData, _ := Compress(data, CompressTypeBrotli)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeBrotli)
	}
}

func TestCompressionRatio(t *testing.T) {
	data := getTestData(&testing.B{})

	compressTypes := map[string]Type{
		"zlib":   CompressTypeZlib,
		"snappy": CompressTypeSnappy,
		"lz4":    CompressTypeLz4,
		"zstd":   CompressTypeZstd,
		"brotli": CompressTypeBrotli,
	}

	for name, tp := range compressTypes {
		t.Run(name, func(t *testing.T) {
			compressed, err := Compress(data, tp)
			assert.NoError(t, err)

			originalSize := len(data)
			compressedSize := len(compressed)

			ratio := float64(compressedSize) / float64(originalSize)
			t.Logf("Algorithm: %v, Original size: %d, Compressed size: %d, Compression ratio: %.2f%%",
				name, originalSize, compressedSize, ratio*100)
		})
	}
}

func getTestData(b *testing.B) []byte {
	type Demo struct {
		Name  string
		Email string
		ID    int
		Age   int
	}
	var buf bytes.Buffer
	for i := 0; i < 50; i++ {
		d := &Demo{
			ID:    i,
			Name:  "user",
			Email: "user@example.com",
			Age:   20 + i%10,
		}
		bin, _ := json.Marshal(d)
		buf.Write(bin)
	}
	b.SetBytes(int64(buf.Len()))
	return buf.Bytes()
}
