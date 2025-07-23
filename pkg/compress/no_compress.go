// Package compress provides compression and decompression utilities.
package compress

// NoCompressor implements Compressor interface.
type NoCompressor struct{}

// Compress compresses data.
func (NoCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

// Decompress decompresses data.
func (NoCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}
