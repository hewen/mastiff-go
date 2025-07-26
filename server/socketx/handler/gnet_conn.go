// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"github.com/panjf2000/gnet/v2"
)

// GnetConn is a wrapper around gnet.Conn that implements the Conn interface.
type GnetConn struct {
	gnet.Conn
}

// AsyncWrite writes data to the connection asynchronously.
func (g *GnetConn) AsyncWrite(data []byte, callback AsyncCallback) error {
	return g.Conn.AsyncWrite(data, func(c gnet.Conn, err error) error {
		return callback(&GnetConn{c}, err)
	})
}
