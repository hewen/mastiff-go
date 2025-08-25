// Package rpcx provides a unified RPC abstraction over gRPC and Connect.
package rpcx

import (
	"sync"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/rpcx/handler"
)

// RPCServer is a server that provides a unified RPC abstraction over gRPC and Connect.
type RPCServer struct {
	handler handler.RPCHandler
	logger  logger.Logger
	mu      sync.Mutex
}

// NewRPCServer creates a new RPCServer.
func NewRPCServer(conf *serverconf.RPCConfig, params handler.RPCBuildParams) (*RPCServer, error) {
	h, err := handler.NewHandler(conf, params)
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
		s.logger.Panicf("rpc server start failed: %v", err)
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
