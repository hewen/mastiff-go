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
	// HTTPTimeoutReadDefault is the default timeout for reading requests.
	HTTPTimeoutReadDefault = 10
	// HTTPTimeoutWriteDefault is the default timeout for writing responses.
	HTTPTimeoutWriteDefault = 10
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
