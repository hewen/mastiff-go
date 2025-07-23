package compress

import (
	"errors"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZstdRoundTrip(t *testing.T) {
	c := ZstdCompressor{}
	plain := []byte("hello zstd")

	compressed, err := c.Compress(plain)
	require.NoError(t, err)
	assert.Greater(t, len(compressed), 0)

	decompressed, err := c.Decompress(compressed)
	require.NoError(t, err)
	assert.Equal(t, plain, decompressed)
}

func TestDecompress_InvalidData(t *testing.T) {
	c := ZstdCompressor{}
	_, err := c.Decompress([]byte("not-zstd-data"))
	assert.Error(t, err)
}

func TestCompress_NewWriterFail(t *testing.T) {
	t.Run("mock writer fail", func(t *testing.T) {
		oldNewWriter := zstdNewWriter
		defer func() { zstdNewWriter = oldNewWriter }()

		zstdNewWriter = func(_ io.Writer, _ ...zstd.EOption) (*zstd.Encoder, error) {
			return nil, errors.New("mock new writer fail")
		}

		c := ZstdCompressor{}
		_, err := c.Compress([]byte("data"))
		assert.EqualError(t, err, "mock new writer fail")
	})
}

func TestDecompress_NewReaderFail(t *testing.T) {
	t.Run("mock reader fail", func(t *testing.T) {
		oldNewReader := zstdNewReader
		defer func() { zstdNewReader = oldNewReader }()

		zstdNewReader = func(_ io.Reader, _ ...zstd.DOption) (*zstd.Decoder, error) {
			return nil, errors.New("mock new reader fail")
		}

		c := ZstdCompressor{}
		_, err := c.Decompress([]byte{0x28, 0xb5, 0x2f, 0xfd})
		assert.EqualError(t, err, "mock new reader fail")
	})
}
