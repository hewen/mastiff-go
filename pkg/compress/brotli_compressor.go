// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"errors"
	"io"

	"github.com/google/brotli/go/cbrotli"
)

// BrotliCompressor implements Compressor interface.
type BrotliCompressor struct{}

// Compress uses Brotli to compress data.
func (BrotliCompressor) Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := cbrotli.NewWriter(&b, cbrotli.WriterOptions{Quality: 5})
	_, writeErr := w.Write(data)
	closeErr := w.Close()
	if writeErr != nil || closeErr != nil {
		return nil, errors.Join(writeErr, closeErr)
	}

	return b.Bytes(), nil
}

// Decompress uses Brotli to decompress data.
func (BrotliCompressor) Decompress(data []byte) ([]byte, error) {
	r := cbrotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}
