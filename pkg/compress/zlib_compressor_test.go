package compress

import (
	"compress/flate"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZlibCompressor_CompressAndDecompress(t *testing.T) {
	c := ZlibCompressor{}
	original := []byte("hello zlib compression world")

	compressed, err := c.Compress(original)
	assert.NoError(t, err, "compress should not return error")
	assert.NotEmpty(t, compressed, "compressed data should not be empty")

	decompressed, err := c.Decompress(compressed)
	assert.NoError(t, err, "decompress should not return error")
	assert.Equal(t, original, decompressed, "decompressed data should match original")
}

func TestZlibCompressor_Compress_NewWriterError(t *testing.T) {
	original := flateNewWriter
	defer func() { flateNewWriter = original }()

	flateNewWriter = func(_ io.Writer, _ int) (*flate.Writer, error) {
		return nil, errors.New("mock new writer error")
	}

	c := ZlibCompressor{}
	_, err := c.Compress([]byte("data"))
	assert.Error(t, err)
	assert.EqualError(t, err, "mock new writer error")
}

func TestZlibCompressor_Decompress_InvalidData(t *testing.T) {
	c := ZlibCompressor{}
	invalid := []byte("this-is-not-zlib-data")

	out, err := c.Decompress(invalid)
	assert.Error(t, err, "should return error on invalid data")
	assert.Nil(t, out, "output should be nil on error")
}
