// Package socketx provides a unified socket abstraction over gnet.
package socketx

import (
	"sync"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/socketx/handler"
)

// SocketServer is a server that provides a unified socket abstraction over gnet.
type SocketServer struct {
	handler handler.SocketHandler
	logger  logger.Logger
	mu      sync.Mutex
}

// NewSocketServer creates a new SocketServer.
func NewSocketServer(conf *serverconf.SocketConfig, params handler.BuildParams) (*SocketServer, error) {
	h, err := handler.NewHandler(conf, params)
	if err != nil {
		return nil, err
	}
	return &SocketServer{
		handler: h,
		logger:  logger.NewLogger(),
	}, nil
}

// Start starts the SocketServer.
func (s *SocketServer) Start() {
	if err := s.handler.Start(); err != nil {
		s.logger.Errorf("socket server start failed: %v", err)
	}
}

// Stop stops the SocketServer.
func (s *SocketServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handler != nil {
		_ = s.handler.Stop()
	}
}

// Name returns the name of the SocketServer.
func (s *SocketServer) Name() string {
	return s.handler.Name()
}

// WithLogger sets the logger for the SocketServer.
func (s *SocketServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
