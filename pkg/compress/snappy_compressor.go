// Package compress provides compression and decompression utilities.
package compress

import "github.com/golang/snappy"

// SnappyCompressor implements Compressor interface.
type SnappyCompressor struct{}

// Compress compresses data.
func (SnappyCompressor) Compress(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

// Decompress decompresses data.
func (SnappyCompressor) Decompress(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}
