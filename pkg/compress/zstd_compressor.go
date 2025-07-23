// Package compress provides compression and decompression utilities.
package compress

import "github.com/klauspost/compress/zstd"

var (
	// zstdNewWriter is a variable to allow for mocking in tests.
	zstdNewWriter = zstd.NewWriter
	// zstdNewReader is a variable to allow for mocking in tests.
	zstdNewReader = zstd.NewReader
)

// ZstdCompressor implements Compressor interface.
type ZstdCompressor struct{}

// Compress compresses data.
func (ZstdCompressor) Compress(data []byte) ([]byte, error) {
	enc, err := zstdNewWriter(nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = enc.Close() }()
	return enc.EncodeAll(data, nil), nil
}

// Decompress decompresses data.
func (ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	dec, err := zstdNewReader(nil)
	if err != nil {
		return nil, err
	}
	defer dec.Close()
	return dec.DecodeAll(data, nil)
}
