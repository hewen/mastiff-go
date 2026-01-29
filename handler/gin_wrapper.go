// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// WrapHandlerGin wraps a handler function into a Gin handler.
func WrapHandlerGin[T any, R any](handle WrapHandlerFunc[T, R]) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &unicontext.GinContext{
			Ctx: c,
		}
		_ = WrapHandler(handle)(ctx)
	}
}
