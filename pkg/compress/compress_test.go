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

func TestCompressTypes(t *testing.T) {
	compressTypes := map[string]Type{
		"no_compress": CompressTypeNoCompress,
		"zlib":        CompressTypeZlib,
		"snappy":      CompressTypeSnappy,
		"lz4":         CompressTypeLz4,
		"zstd":        CompressTypeZstd,
		"brotli":      CompressTypeBrotli,
	}

	for name, tp := range compressTypes {
		t.Run(name, func(t *testing.T) {
			data := []byte(strings.Repeat("test", 100))
			compressData, err := Compress(data, tp)
			assert.Nil(t, err)

			decompressData, err := Decompress(compressData, tp)
			assert.Equal(t, decompressData, data)
			assert.Nil(t, err)
		})
	}
}
