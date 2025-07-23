// Package compress provides compression and decompression utilities.
package compress

import "github.com/pierrec/lz4/v4"

// Lz4Compressor implements Compressor interface.
type Lz4Compressor struct{}

// Compress compresses data.
func (Lz4Compressor) Compress(data []byte) ([]byte, error) {
	buf := make([]byte, lz4.CompressBlockBound(len(data)))
	var c lz4.Compressor
	n, err := c.CompressBlock(data, buf)
	return buf[:n], err
}

// Decompress decompresses data.
func (Lz4Compressor) Decompress(data []byte) ([]byte, error) {
	out := make([]byte, len(data)*20)
	n, err := lz4.UncompressBlock(data, out)
	if err != nil {
		return nil, err
	}
	return out[:n], nil
}
