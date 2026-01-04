// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"context"

	"github.com/hewen/mastiff-go/server/httpx/handler"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// httpxContext implements the Context interface for Httpx.
type httpxContext struct {
	Ctx unicontext.UniversalContext
}

// JSON implements the Context interface.
func (g *httpxContext) JSON(code int, obj any) error {
	return g.Ctx.JSON(code, obj)
}

// BindJSON implements the Context interface.
func (g *httpxContext) BindJSON(obj any) error {
	return g.Ctx.BindJSON(obj)
}

// Set implements the Context interface.
func (g *httpxContext) Set(key string, val any) {
	g.Ctx.Set(key, val)
}

// RequestContext implements the Context interface.
func (g *httpxContext) RequestContext() context.Context {
	return g.Ctx.Request().Context()
}

// WrapHandlerHttpx wraps a handler function into a Httpx handler.
func WrapHandlerHttpx[T any, R any](handle WrapHandlerFunc[T, R]) handler.HTTPHandlerFunc {
	return func(c unicontext.UniversalContext) error {
		return WrapHandler(handle)(&httpxContext{Ctx: c})
	}
}
