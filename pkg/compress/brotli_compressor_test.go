package compress

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrotliCompressor_CompressDecompress(t *testing.T) {
	c := NewBrotliCompressor()

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
	c := NewBrotliCompressor()

	invalidData := []byte("not valid brotli data")
	decompressed, err := c.Decompress(invalidData)
	assert.Error(t, err)
	assert.Equal(t, 0, len(decompressed))
}

func TestBrotliCompressor_Compress_EmptyInput(t *testing.T) {
	c := NewBrotliCompressor()

	compressed, err := c.Compress(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, compressed)

	decompressed, err := c.Decompress(compressed)
	assert.NoError(t, err)
	assert.Empty(t, decompressed)
}

func TestBrotliCompressor_CompressWriteError(t *testing.T) {
	c := &BrotliCompressor{
		writerFactory: func(_ io.Writer) io.WriteCloser {
			return &errorWriter{}
		},
	}
	_, err := c.Compress([]byte("fail"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func (e *errorWriter) Close() error {
	return errors.New("close error")
}
