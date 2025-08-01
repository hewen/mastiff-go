// Package handler provides a unified HTTP abstraction over Gin and Fiber.
package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

var (
	// ErrEmptyHTTPConf is the error returned when the HTTP config is empty.
	ErrEmptyHTTPConf = errors.New("http config is empty")
)

const (
	// HTTPTimeoutDefault is the default timeout for reading requests.
	HTTPTimeoutDefault = 10
)

// HTTPHandlerFunc is the function signature for HTTP handlers.
type HTTPHandlerFunc func(unicontext.UniversalContext) error

// HTTPHandler represents a unified HTTP server.
type HTTPHandler interface {
	RouterGroup

	Start() error
	Stop() error
	Name() string
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
}

// RouterGroup represents a group of routes.
type RouterGroup interface {
	Router
	Group(relativePath string, handlers ...HTTPHandlerFunc) RouterGroup
}

// Router represents a router.
type Router interface {
	Use(...HTTPHandlerFunc) Router
	Handle(string, string, ...HTTPHandlerFunc) Router
	Any(string, ...HTTPHandlerFunc) Router
	Get(string, ...HTTPHandlerFunc) Router
	Post(string, ...HTTPHandlerFunc) Router
	Delete(string, ...HTTPHandlerFunc) Router
	Patch(string, ...HTTPHandlerFunc) Router
	Put(string, ...HTTPHandlerFunc) Router
	Options(string, ...HTTPHandlerFunc) Router
	Head(string, ...HTTPHandlerFunc) Router
	Match(methods []string, path string, handlers ...HTTPHandlerFunc) Router
}

// toDuration converts a timeout in seconds to a time.Duration.
func toDuration(sec int64) time.Duration {
	if sec == 0 {
		return HTTPTimeoutDefault * time.Second
	}
	return time.Duration(sec) * time.Second
}
