// Package httpx provides a unified HTTP abstraction over Gin and Fiber.
package httpx

import (
	"errors"
)

var (
	// ErrEmptyHTTPConf is the error returned when the HTTP config is empty.
	ErrEmptyHTTPConf = errors.New("http config is empty")
)

const (
	// HTTPTimeoutDefault is the default timeout for reading requests.
	HTTPTimeoutDefault = 10
)

// HTTPHandler represents a unified HTTP server.
type HTTPHandler interface {
	Start() error
	Stop() error
	Name() string
}

// HTTPHandlerBuilder builds a specific HTTPHandler (Gin, Fiber, or Std).
type HTTPHandlerBuilder interface {
	// BuildHandler builds a specific HTTPHandler (Gin, Fiber, or Std).
	BuildHandler() (HTTPHandler, error)
}
