package compress

import (
	"testing"

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
