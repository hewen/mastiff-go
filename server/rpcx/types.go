// Package rpcx provides a unified RPC abstraction over gRPC and Connect.
package rpcx

const (
	// RPCTimeoutDefault is the default timeout for RPC requests.
	RPCTimeoutDefault = 10
)

// RPCHandler represents a pluggable RPC backend (gRPC or Connect).
type RPCHandler interface {
	Start() error
	Stop() error
	Name() string
}

// RPCHandlerBuilder builds a specific RPCHandler (gRPC or Connect).
type RPCHandlerBuilder interface {
	BuildRPC() (RPCHandler, error)
}
