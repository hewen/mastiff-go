// Package httpx provides a unified HTTP abstraction over Gin.
package httpx

import "net/http"

// HTTPHandlerBuilder builds a specific HTTPHandler (Gin, Fiber, or Std).
type HTTPHandlerBuilder interface {
	// BuildHandler builds a specific HTTPHandler (Gin, Fiber, or Std).
	BuildHandler() http.Handler
}
