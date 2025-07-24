// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"errors"
	"io"

	"github.com/andybalholm/brotli"
)

// BrotliCompressor implements Compressor interface.
type BrotliCompressor struct{}

// Compress uses Brotli to compress data.
func (BrotliCompressor) Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := brotli.NewWriter(&b)
	_, writeErr := w.Write(data)
	closeErr := w.Close()
	if writeErr != nil || closeErr != nil {
		return nil, errors.Join(writeErr, closeErr)
	}

	return b.Bytes(), nil
}

// Decompress uses Brotli to decompress data.
func (BrotliCompressor) Decompress(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}
