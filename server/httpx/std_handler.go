// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
)

// StdHTTPHandlerBuilder builds a standard HTTP handler.
type StdHTTPHandlerBuilder struct {
	Handler http.Handler
	Conf    *serverconf.HTTPConfig
}

// BuildHandler builds a standard HTTP handler.
func (s *StdHTTPHandlerBuilder) BuildHandler() (HTTPHandler, error) {
	if s.Conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	return &StdHandler{
		server: http.Server{
			Addr:         s.Conf.Addr,
			Handler:      s.Handler,
			ReadTimeout:  time.Duration(s.Conf.ReadTimeout),
			WriteTimeout: time.Duration(s.Conf.WriteTimeout),
		},
		name: "std",
	}, nil
}

// StdHandler is a handler that provides a unified HTTP abstraction over standard HTTP.
type StdHandler struct {
	name   string
	server http.Server
}

// Start starts the StdHandler.
func (s *StdHandler) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the StdHandler.
func (s *StdHandler) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Name returns the name of the StdHandler.
func (s *StdHandler) Name() string {
	return fmt.Sprintf("http %s server(%s)", s.name, s.server.Addr)
}
