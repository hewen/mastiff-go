package compress

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCompressorErrorType(t *testing.T) {
	_, err := GetCompressor(999)
	assert.Error(t, err)
}

func TestCompressErrorType(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	_, err := Compress(data, 999)
	assert.Error(t, err)

	_, err = Decompress(data, 999)
	assert.Error(t, err)
}

func TestCompressNoCompress(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	compressData, err := Compress(data, CompressTypeNoCompress)
	assert.Nil(t, err)
	assert.Equal(t, compressData, data)

	decompressData, err := Decompress(compressData, CompressTypeNoCompress)
	assert.Equal(t, decompressData, data)
	assert.Nil(t, err)
}

func TestCompressZlibCompress(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	compressData, err := Compress(data, CompressTypeZlib)
	assert.Nil(t, err)

	decompressData, err := Decompress(compressData, CompressTypeZlib)
	assert.Equal(t, decompressData, data)
	assert.Nil(t, err)
}

func TestCompressSnappyCompress(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	compressData, err := Compress(data, CompressTypeSnappy)
	assert.Nil(t, err)

	decompressData, err := Decompress(compressData, CompressTypeSnappy)
	assert.Equal(t, decompressData, data)
	assert.Nil(t, err)
}

func TestCompressLz4Compress(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	compressData, err := Compress(data, CompressTypeLz4)

	assert.Nil(t, err)

	decompressData, err := Decompress(compressData, CompressTypeLz4)
	assert.Equal(t, decompressData, data)
	assert.Nil(t, err)
}

func TestCompressZstdCompress(t *testing.T) {
	data := []byte(strings.Repeat("test", 100))
	compressData, err := Compress(data, CompressTypeZstd)
	assert.Nil(t, err)

	decompressData, err := Decompress(compressData, CompressTypeZstd)
	assert.Equal(t, decompressData, data)
	assert.Nil(t, err)
}

func BenchmarkCompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeZlib)
	}
}

func BenchmarkDecompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	compressData, _ := Compress(data, CompressTypeZlib)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeZlib)
	}
}

func BenchmarkSnappyCompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeSnappy)
	}
}

func BenchmarkSnappyDecompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	compressData, _ := Compress(data, CompressTypeSnappy)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeSnappy)
	}
}

func BenchmarkLz4Compress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeLz4)
	}
}

func BenchmarkLz4Decompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	compressData, _ := Compress(data, CompressTypeLz4)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeLz4)
	}
}

func BenchmarkZstdCompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressTypeZstd)
	}
}

func BenchmarkZstdDecompress(b *testing.B) {
	data := []byte(strings.Repeat("test", 100))
	compressData, _ := Compress(data, CompressTypeZstd)

	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressData, CompressTypeZstd)
	}
}
