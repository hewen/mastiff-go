// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// WrapHandlerFiber wraps a handler function into a Fiber handler.
func WrapHandlerFiber[T any, R any](handle WrapHandlerFunc[T, R]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &unicontext.FiberContext{
			Ctx: c,
		}
		return WrapHandler(handle)(ctx)
	}
}
