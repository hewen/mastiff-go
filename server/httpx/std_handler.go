// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import "net/http"

// StdHTTPHandlerBuilder builds a standard HTTP handler.
type StdHTTPHandlerBuilder struct {
	Handler http.Handler
}

// BuildHandler builds a standard HTTP handler.
func (s *StdHTTPHandlerBuilder) BuildHandler() http.Handler {
	return s.Handler
}
