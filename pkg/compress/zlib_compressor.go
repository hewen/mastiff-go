// Package compress provides compression and decompression utilities.
package compress

import (
	"bytes"
	"compress/flate"
	"io"
)

// flateNewWriter is a variable to allow for mocking in tests.
var flateNewWriter = flate.NewWriter

// ZlibCompressor implements Compressor interface.
type ZlibCompressor struct{}

// Compress compresses data.
func (ZlibCompressor) Compress(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	w, err := flateNewWriter(b, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
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
