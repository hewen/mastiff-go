// Package compress provides compression and decompression utilities.
package compress

// Type is the type of compression.
type Type uint16

const (
	// CompressTypeNoCompress is the type of no compression.
	CompressTypeNoCompress Type = iota
	// CompressTypeZlib is the type of zlib compression.
	CompressTypeZlib
	// CompressTypeSnappy is the type of snappy compression.
	CompressTypeSnappy
	// CompressTypeLz4 is the type of lz4 compression.
	CompressTypeLz4
	// CompressTypeZstd is the type of zstd compression.
	CompressTypeZstd
)

// Compressor is an interface for compressing and decompressing data.
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}
