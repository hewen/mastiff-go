// Package compress provides compression and decompression utilities.
package compress

import "github.com/klauspost/compress/zstd"

var (
	// zstdEncoder is a variable to allow for mocking in tests.
	zstdEncoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	// zstdDecoder is a variable to allow for mocking in tests.
	zstdDecoder, _ = zstd.NewReader(nil)
)

// ZstdCompressor implements Compressor interface.
type ZstdCompressor struct{}

// Compress compresses data.
func (ZstdCompressor) Compress(data []byte) ([]byte, error) {
	return zstdEncoder.EncodeAll(data, nil), nil
}

// Decompress decompresses data.
func (ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	return zstdDecoder.DecodeAll(data, nil)
}
