package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// AsFiberHandler converts a list of UniversalHandlerFunc to a list of fiber.Handler.
func AsFiberHandler(handlers ...UniversalHandlerFunc) []fiber.Handler {
	out := make([]fiber.Handler, len(handlers))
	for i, h := range handlers {
		handler := h
		out[i] = func(c *fiber.Ctx) error {
			ctx := &unicontext.FiberContext{Ctx: c}
			return handler(ctx)
		}
	}
	return out
}

// AsGinHandler converts a list of UniversalHandlerFunc to a list of gin.HandlerFunc.
func AsGinHandler(handlers ...UniversalHandlerFunc) []gin.HandlerFunc {
	out := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		handler := h
		out[i] = func(c *gin.Context) {
			ctx := &unicontext.GinContext{Ctx: c}
			_ = handler(ctx)
		}
	}
	return out
}

// FromHTTPHandlerFunc converts a HTTP handler function to a UniversalHandlerFunc.
func FromHTTPHandlerFunc(h func(w http.ResponseWriter, r *http.Request)) UniversalHandlerFunc {
	return func(ctx unicontext.UniversalContext) error {
		h(ctx.ResponseWriter(), ctx.Request())
		return nil
	}
}

// FromHTTPHandler converts a HTTP handler to a UniversalHandlerFunc.
func FromHTTPHandler(h http.Handler) UniversalHandlerFunc {
	return func(ctx unicontext.UniversalContext) error {
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
		return nil
	}
}
