// Package handler provides a unified socket abstraction over gnet.
package handler

// SocketHandler represents the top-level abstraction for a socket server.
type SocketHandler interface {
	// Start starts the socket server.
	Start() error

	// Stop stops the socket server gracefully.
	Stop() error

	// Name returns the server name or protocol identifier.
	Name() string
}
