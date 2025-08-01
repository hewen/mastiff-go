// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"errors"
	"io"

	"github.com/andybalholm/brotli"
)

// BrotliCompressor implements Compressor interface.
type BrotliCompressor struct {
	writerFactory func(io.Writer) io.WriteCloser
}

// NewBrotliCompressor create BrotliCompressor.
func NewBrotliCompressor() *BrotliCompressor {
	return &BrotliCompressor{
		writerFactory: func(w io.Writer) io.WriteCloser {
			return brotli.NewWriter(w)
		},
	}
}

// Compress uses Brotli to compress data.
func (b BrotliCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := b.writerFactory(&buf)
	_, writeErr := w.Write(data)
	closeErr := w.Close()
	if writeErr != nil || closeErr != nil {
		return nil, errors.Join(writeErr, closeErr)
	}

	return buf.Bytes(), nil
}

// Decompress uses Brotli to decompress data.
func (BrotliCompressor) Decompress(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}
