package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrotliCompressor_CompressDecompress(t *testing.T) {
	c := BrotliCompressor{}

	origin := []byte("This is a test string for Brotli compression.")

	compressed, err := c.Compress(origin)
	assert.NoError(t, err)
	assert.NotEmpty(t, compressed)
	assert.NotEqual(t, origin, compressed, "compressed data should differ from original")

	decompressed, err := c.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, origin, decompressed, "decompressed data should equal original")
}

func TestBrotliCompressor_Decompress_InvalidData(t *testing.T) {
	c := BrotliCompressor{}

	invalidData := []byte("not valid brotli data")
	decompressed, err := c.Decompress(invalidData)
	assert.Error(t, err)
	assert.Equal(t, 0, len(decompressed))
}

func TestBrotliCompressor_Compress_EmptyInput(t *testing.T) {
	c := BrotliCompressor{}

	compressed, err := c.Compress([]byte{})
	assert.NoError(t, err)
	assert.NotEmpty(t, compressed)

	decompressed, err := c.Decompress(compressed)
	assert.NoError(t, err)
	assert.Empty(t, decompressed)
}
