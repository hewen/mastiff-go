// Package handler provides a unified RPC abstraction over gRPC and Connect.
package handler

import (
	"fmt"
)

const (
	// RPCTimeoutDefault is the default timeout for RPC requests.
	RPCTimeoutDefault = 10
)

// ErrEmptyRPCConf is the error returned when the RPC config is empty.
var ErrEmptyRPCConf = fmt.Errorf("empty rpc config")

// RPCHandler represents a pluggable RPC backend (gRPC or Connect).
type RPCHandler interface {
	Start() error
	Stop() error
	Name() string
}
