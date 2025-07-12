// Package server http server implementation
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
)

var (
	// ErrEmptyHTTPConf is returned when the HTTP config is empty.
	ErrEmptyHTTPConf = errors.New("not set queue name empty")
)

const (
	// HTTPTimeoutReadDefault is the default read timeout for the HTTP server.
	HTTPTimeoutReadDefault = 10
	// HTTPTimeoutWriteDefault is the default write timeout for the HTTP server.
	HTTPTimeoutWriteDefault = 10
)

// HTTPService defines the configuration for an HTTP server.
type HTTPService struct {
	s      *http.Server
	logger logger.Logger
	addr   string
	mu     sync.Mutex
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(conf *serverconf.HTTPConfig, initRoute func(r *gin.Engine), extraMiddlewares ...gin.HandlerFunc) (*HTTPService, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	if conf.TimeoutRead == 0 {
		conf.TimeoutRead = HTTPTimeoutReadDefault
	}
	if conf.TimeoutWrite == 0 {
		conf.TimeoutWrite = HTTPTimeoutWriteDefault
	}

	handler := NewGinAPIHandler(*conf, initRoute, extraMiddlewares...)
	srv := &http.Server{
		Addr:         conf.Addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.TimeoutRead) * time.Second,
		WriteTimeout: time.Duration(conf.TimeoutWrite) * time.Second,
	}

	service := &HTTPService{
		addr:   conf.Addr,
		s:      srv,
		logger: logger.NewLogger(),
	}

	return service, nil
}

// Start starts the HTTP server.
func (s *HTTPService) Start() {
	if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Errorf("http service failed: %v", err)
	}
}

// Stop gracefully stops the HTTP server.
func (s *HTTPService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.s == nil {
		s.logger = nil
		return
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.s.Shutdown(ctx)
	if err != nil {
		s.logger.Errorf("http server(%s) shutdown error: %v", s.addr, err)
	}
	s.s = nil
}

// Name returns the name of the HTTP server.
func (s *HTTPService) Name() string {
	return fmt.Sprintf("http server(%s)", s.addr)
}

// WithLogger sets the logger for the HTTP server.
func (s *HTTPService) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
