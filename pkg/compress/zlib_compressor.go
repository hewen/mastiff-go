// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
)

type writerCloser interface {
	io.Writer
	io.Closer
}

// flateNewWriter is a variable to allow for mocking in tests.
var flateNewWriter = func(w io.Writer, level int) (writerCloser, error) {
	return flate.NewWriter(w, level)
}

// ZlibCompressor implements Compressor interface.
type ZlibCompressor struct{}

// Compress compresses data.
func (ZlibCompressor) Compress(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	w, err := flateNewWriter(b, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, writeErr := w.Write(data)
	closeErr := w.Close()

	if writeErr != nil || closeErr != nil {
		return nil, errors.Join(writeErr, closeErr)
	}

	return b.Bytes(), nil
}

// Decompress decompresses data.
func (ZlibCompressor) Decompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r := flate.NewReader(b)
	d := []byte{}
	var err error
	for {
		buf := make([]byte, 1024)
		var n int
		n, err = r.Read(buf)
		if n > 0 {
			d = append(d, buf[0:n]...)
		} else {
			break
		}
		if err != nil {
			break
		}

	}

	if err != io.EOF {
		return nil, err
	}

	err = r.Close()
	return d, err
}
