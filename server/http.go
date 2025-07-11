// Package server http server implementation
package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
)

var (
	// ErrEmptyHTTPConf is returned when the HTTP config is empty.
	ErrEmptyHTTPConf = errors.New("empty http config")
)

const (
	// HTTPTimeoutReadDefault is the default read timeout for the HTTP server.
	HTTPTimeoutReadDefault = 10
	// HTTPTimeoutWriteDefault is the default write timeout for the HTTP server.
	HTTPTimeoutWriteDefault = 10
)

// HTTPService defines the configuration for an HTTP server.
type HTTPService struct {
	s    *http.Server
	l    logger.Logger
	addr string
	mu   sync.Mutex
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(conf *HTTPConf, initRoute func(r *gin.Engine), extraMiddlewares ...gin.HandlerFunc) (*HTTPService, error) {
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
		addr: conf.Addr,
		s:    srv,
		l:    logger.NewLogger(),
	}

	return service, nil
}

// Start starts the HTTP server.
func (s *HTTPService) Start() {
	s.l.Infof("Start http service %s", s.addr)
	if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.l.Errorf("http service failed: %v", err)
	}
}

// Stop gracefully stops the HTTP server.
func (s *HTTPService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.l.Infof("Shutdown service %s", s.addr)
	if s.s == nil {
		return
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.s.Shutdown(ctx)
	if err != nil {
		s.l.Errorf("Error during server shutdown: %v", err)
	}
	s.s = nil
}
