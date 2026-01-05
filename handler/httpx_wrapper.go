// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"github.com/hewen/mastiff-go/server/httpx/handler"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// WrapHandlerHttpx wraps a handler function into a Httpx handler.
func WrapHandlerHttpx[T any, R any](handle WrapHandlerFunc[T, R]) handler.HTTPHandlerFunc {
	return func(c unicontext.UniversalContext) error {
		return WrapHandler(handle)(c)
	}
}
