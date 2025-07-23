package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLz4Compressor_CompressAndDecompress(t *testing.T) {
	c := Lz4Compressor{}
	original := []byte("this is lz4 test data...")

	compressed, err := c.Compress(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, compressed)

	decompressed, err := c.Decompress(compressed)
	assert.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestLz4Compressor_Decompress_InvalidData(t *testing.T) {
	c := Lz4Compressor{}
	invalid := []byte("not-valid-lz4-data")

	out, err := c.Decompress(invalid)
	assert.Error(t, err)
	assert.Nil(t, out)
}
