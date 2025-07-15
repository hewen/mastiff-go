// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import (
	"sync"

	"github.com/hewen/mastiff-go/logger"
)

// HTTPServer is a server that provides a unified HTTP abstraction over Gin.
type HTTPServer struct {
	handler HTTPHandler
	logger  logger.Logger
	mu      sync.Mutex
}

// NewHTTPServer creates a new HTTPServer.
func NewHTTPServer(builder HTTPHandlerBuilder) (*HTTPServer, error) {
	h, err := builder.BuildHandler()
	if err != nil {
		return nil, err
	}

	return &HTTPServer{
		handler: h,
		logger:  logger.NewLogger(),
	}, nil
}

// Start starts the HTTPServer.
func (s *HTTPServer) Start() {
	if err := s.handler.Start(); err != nil {
		s.logger.Errorf("http server start failed: %v", err)
	}

}

// Stop stops the HTTPServer.
func (s *HTTPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		_ = s.handler.Stop()
	}
}

// Name returns the name of the HTTPServer.
func (s *HTTPServer) Name() string {
	return s.handler.Name()
}

// WithLogger sets the logger for the HTTPServer.
func (s *HTTPServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
