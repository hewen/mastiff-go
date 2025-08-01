// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ginContext implements the Context interface for Gin.
type ginContext struct {
	Context *gin.Context
}

// JSON implements the Context interface.
func (g *ginContext) JSON(code int, obj any) error {
	g.Context.JSON(code, obj)
	return nil
}

// BindJSON implements the Context interface.
func (g *ginContext) BindJSON(obj any) error {
	return g.Context.ShouldBindJSON(obj)
}

// Set implements the Context interface.
func (g *ginContext) Set(key string, val any) {
	g.Context.Set(key, val)
}

// RequestContext implements the Context interface.
func (g *ginContext) RequestContext() context.Context {
	return g.Context.Request.Context()
}

// WrapHandlerGin wraps a handler function into a Gin handler.
func WrapHandlerGin[T any, R any](handle WrapHandlerFunc[T, R]) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = WrapHandler(handle)(&ginContext{Context: c})
	}
}
