// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
)

// HTTPServer is a server that provides a unified HTTP abstraction over Gin.
type HTTPServer struct {
	s      *http.Server
	logger logger.Logger
	addr   string
	mu     sync.Mutex
}

// NewHTTPServer creates a new HTTPServer.
func NewHTTPServer(conf *serverconf.HTTPConfig, builder HTTPHandlerBuilder) (*HTTPServer, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}
	if conf.TimeoutRead == 0 {
		conf.TimeoutRead = HTTPTimeoutReadDefault
	}
	if conf.TimeoutWrite == 0 {
		conf.TimeoutWrite = HTTPTimeoutWriteDefault
	}

	srv := &http.Server{
		Addr:         conf.Addr,
		Handler:      builder.BuildHandler(),
		ReadTimeout:  time.Duration(conf.TimeoutRead) * time.Second,
		WriteTimeout: time.Duration(conf.TimeoutWrite) * time.Second,
	}

	return &HTTPServer{
		addr:   conf.Addr,
		s:      srv,
		logger: logger.NewLogger(),
	}, nil
}

// Start starts the HTTPServer.
func (s *HTTPServer) Start() {
	if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Errorf("http server start failed: %v", err)
	}
}

// Stop stops the HTTPServer.
func (s *HTTPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.s == nil {
		s.logger = nil
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.s.Shutdown(ctx); err != nil {
		s.logger.Errorf("http server(%s) shutdown error: %v", s.addr, err)
	}
	s.s = nil
}

// Name returns the name of the HTTPServer.
func (s *HTTPServer) Name() string {
	return fmt.Sprintf("http server(%s)", s.addr)
}

// WithLogger sets the logger for the HTTPServer.
func (s *HTTPServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
