// Package rpcx provides a unified RPC abstraction over gRPC and Connect.
package rpcx

import (
	"fmt"
	"sync"

	"github.com/hewen/mastiff-go/logger"
)

// ErrEmptyRPCConf is the error returned when the RPC config is empty.
var ErrEmptyRPCConf = fmt.Errorf("empty rpc config")

// RPCServer is a server that provides a unified RPC abstraction over gRPC and Connect.
type RPCServer struct {
	handler RPCHandler
	logger  logger.Logger
	mu      sync.Mutex
}

// NewRPCServer creates a new RPCServer.
func NewRPCServer(builder RPCHandlerBuilder) (*RPCServer, error) {
	h, err := builder.BuildRPC()
	if err != nil {
		return nil, err
	}

	return &RPCServer{
		handler: h,
		logger:  logger.NewLogger(),
	}, nil
}

// Start starts the RPCServer.
func (s *RPCServer) Start() {
	if err := s.handler.Start(); err != nil {
		s.logger.Errorf("rpc server start failed: %v", err)
	}
}

// Stop stops the RPCServer.
func (s *RPCServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		_ = s.handler.Stop()
	}
}

// Name returns the name of the RPCServer.
func (s *RPCServer) Name() string {
	return s.handler.Name()
}

// WithLogger sets the logger for the RPCServer.
func (s *RPCServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
