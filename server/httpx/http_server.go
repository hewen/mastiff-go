// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import (
	"sync"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/httpx/handler"
)

// HTTPServer is a server that provides a unified HTTP abstraction over Gin.
type HTTPServer struct {
	handler.HTTPHandler
	logger logger.Logger
	mu     sync.Mutex
}

// NewHTTPServer creates a new HTTPServer.
func NewHTTPServer(conf *serverconf.HTTPConfig, opts ...handler.ServerOption) (*HTTPServer, error) {
	h, err := handler.NewHandler(conf, opts...)
	if err != nil {
		return nil, err
	}

	return &HTTPServer{
		HTTPHandler: h,
		logger:      logger.NewLogger(),
	}, nil

}

// Start starts the HTTPServer.
func (s *HTTPServer) Start() {
	if err := s.HTTPHandler.Start(); err != nil {
		s.logger.Errorf("http server start failed: %v", err)
	}

}

// Stop stops the HTTPServer.
func (s *HTTPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.HTTPHandler != nil {
		_ = s.HTTPHandler.Stop()
	}
}

// Name returns the name of the HTTPServer.
func (s *HTTPServer) Name() string {
	return s.HTTPHandler.Name()
}

// WithLogger sets the logger for the HTTPServer.
func (s *HTTPServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
