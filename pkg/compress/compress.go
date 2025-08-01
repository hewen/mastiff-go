// Package compress provides compression and decompression utilities.
package compress

import "errors"

func init() {
	RegisterCompressor(CompressTypeNoCompress, NoCompressor{})
	RegisterCompressor(CompressTypeZlib, ZlibCompressor{})
	RegisterCompressor(CompressTypeSnappy, SnappyCompressor{})
	RegisterCompressor(CompressTypeLz4, Lz4Compressor{})
	RegisterCompressor(CompressTypeZstd, ZstdCompressor{})
	RegisterCompressor(CompressTypeBrotli, NewBrotliCompressor())
}

// compressorRegistry is a map of compress type to compressor.
var compressorRegistry = make(map[Type]Compressor)

// RegisterCompressor registers a compressor.
func RegisterCompressor(tp Type, c Compressor) {
	compressorRegistry[tp] = c
}

// GetCompressor returns a compressor by type.
func GetCompressor(tp Type) (Compressor, error) {
	c, ok := compressorRegistry[tp]
	if !ok {
		return nil, errors.New("compressor not registered")
	}
	return c, nil
}

// Compress compresses data.
func Compress(data []byte, tp Type) ([]byte, error) {
	c, err := GetCompressor(tp)
	if err != nil {
		return nil, err
	}
	return c.Compress(data)
}

// Decompress decompresses data.
func Decompress(data []byte, tp Type) ([]byte, error) {
	c, err := GetCompressor(tp)
	if err != nil {
		return nil, err
	}
	return c.Decompress(data)
}
