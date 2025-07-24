// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/pierrec/lz4/v4"
)

// Lz4Compressor implements Compressor interface.
type Lz4Compressor struct{}

// lz4WriterPool is a pool of lz4 writers.
var lz4WriterPool = sync.Pool{
	New: func() any {
		return lz4.NewWriter(nil)
	},
}

// lz4ReaderPool is a pool of lz4 readers.
var lz4ReaderPool = sync.Pool{
	New: func() any {
		return lz4.NewReader(nil)
	},
}

// Compress compresses data.
func (Lz4Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := lz4WriterPool.Get().(*lz4.Writer)
	defer lz4WriterPool.Put(writer)

	writer.Reset(&buf)
	_, writeErr := writer.Write(data)
	closeErr := writer.Close()
	if writeErr != nil || closeErr != nil {
		return nil, errors.Join(writeErr, closeErr)
	}

	return buf.Bytes(), nil
}

// Decompress decompresses data.
func (Lz4Compressor) Decompress(data []byte) ([]byte, error) {
	reader := lz4ReaderPool.Get().(*lz4.Reader)
	defer lz4ReaderPool.Put(reader)

	reader.Reset(bytes.NewReader(data))
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
