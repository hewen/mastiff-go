package compress

import (
	"compress/flate"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	flateNewWriter = func(_ io.Writer, _ int) (writerCloser, error) {
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

type mockFlateWriter struct {
	writeErr error
	closeErr error
}

func (m *mockFlateWriter) Write(p []byte) (int, error) {
	return len(p), m.writeErr
}

func (m *mockFlateWriter) Close() error {
	return m.closeErr
}

func TestCompress_WriteError(t *testing.T) {
	defer func() {
		flateNewWriter = func(w io.Writer, level int) (writerCloser, error) {
			return flate.NewWriter(w, level)
		}
	}()

	flateNewWriter = func(_ io.Writer, _ int) (writerCloser, error) {
		return &mockFlateWriter{writeErr: errors.New("mock write error")}, nil
	}

	_, err := ZlibCompressor{}.Compress([]byte("data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "mock write error")
}
