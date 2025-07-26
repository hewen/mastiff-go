// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"time"

	"github.com/hewen/mastiff-go/server/socketx/codec"
	"github.com/panjf2000/gnet/v2"
)

// AsyncCallback defines the callback function used after asynchronous writes.
type AsyncCallback func(c Conn, err error) error

// SocketHandler represents the top-level abstraction for a socket server,
// providing lifecycle control and device-based message pushing.
type SocketHandler interface {
	// Start starts the socket server.
	Start() error

	// Stop stops the socket server gracefully.
	Stop() error

	// Name returns the server name or protocol identifier.
	Name() string

	// PushTo sends a message to a specific device by ID, with an optional async callback.
	PushTo(deviceID string, msg codec.Message, callback AsyncCallback) error

	// UnbindDevice unbinds a device by ID.
	UnbindDevice(deviceID string)
}

// Conn abstracts a connection in the socket server.
type Conn interface {
	// AsyncWrite writes data to the connection asynchronously.
	AsyncWrite(data []byte, callback AsyncCallback) error

	// Write writes data to the connection synchronously.
	Write(data []byte) (int, error)

	// Close closes the connection.
	Close() error

	// Context returns the user-defined context bound to the connection.
	Context() any

	// SetContext sets a custom context object on the connection.
	SetContext(ctx any)
}

// GnetEventHandler defines the application-level event hooks
// that are triggered during the lifecycle of socket connections.
// This handler is typically wrapped inside gnet event loop.
type GnetEventHandler interface {
	// OnBoot is called once when the engine starts.
	OnBoot(eng gnet.Engine) gnet.Action

	// OnShutdown is called once when the engine is shutting down.
	OnShutdown(eng gnet.Engine)

	// OnOpen is called when a new connection is opened.
	OnOpen(c gnet.Conn) (out []byte, action gnet.Action)

	// OnClose is called when a connection is closed.
	OnClose(c gnet.Conn, err error) gnet.Action

	// OnTick is called periodically by the event loop.
	OnTick() (delay time.Duration, action gnet.Action)

	// OnHandshakeMessage is triggered when a complete handshake message is received.
	OnHandshakeMessage(c gnet.Conn, data codec.Message) codec.Message

	// OnMessage is triggered when a regular message is received.
	OnMessage(c gnet.Conn, data codec.Message) codec.Message

	// SecureCodecRegistry provides codec negotiation and selection logic.
	SecureCodecRegistry
}

// SecureCodecRegistry allows matching handshake prefixes
// and instantiating protocol-specific codecs.
type SecureCodecRegistry interface {
	// MatchHandshakePrefix returns the protocol name and whether the prefix matches a known secure codec.
	MatchHandshakePrefix(data []byte) (protocol string, matched bool)

	// NewCodec returns a new SecureCodec instance for the given protocol.
	NewCodec(protocol string) (codec.SecureCodec, error)
}
