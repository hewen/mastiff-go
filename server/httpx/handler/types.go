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

// UniversalHandlerFunc is the function signature for HTTP handlers.
type UniversalHandlerFunc func(unicontext.UniversalContext) error

// UniversalHandler represents a unified HTTP server.
type UniversalHandler interface {
	RouterGroup

	Start() error
	Stop() error
	Name() string
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
}

// RouterGroup represents a group of routes.
type RouterGroup interface {
	Router
	Group(relativePath string, handlers ...UniversalHandlerFunc) RouterGroup
}

// Router represents a router.
type Router interface {
	Use(...UniversalHandlerFunc) Router
	Handle(string, string, ...UniversalHandlerFunc) Router
	Any(string, ...UniversalHandlerFunc) Router
	Get(string, ...UniversalHandlerFunc) Router
	Post(string, ...UniversalHandlerFunc) Router
	Delete(string, ...UniversalHandlerFunc) Router
	Patch(string, ...UniversalHandlerFunc) Router
	Put(string, ...UniversalHandlerFunc) Router
	Options(string, ...UniversalHandlerFunc) Router
	Head(string, ...UniversalHandlerFunc) Router
	Match(methods []string, path string, handlers ...UniversalHandlerFunc) Router
}

// toDuration converts a timeout in seconds to a time.Duration.
func toDuration(sec int64) time.Duration {
	if sec == 0 {
		return HTTPTimeoutDefault * time.Second
	}
	return time.Duration(sec) * time.Second
}
