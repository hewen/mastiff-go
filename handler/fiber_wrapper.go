// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// fiberContext implements the Context interface for Fiber.
type fiberContext struct {
	Ctx *fiber.Ctx
}

// JSON implements the Context interface.
func (f *fiberContext) JSON(code int, obj any) error {
	return f.Ctx.Status(code).JSON(obj)
}

// BindJSON implements the Context interface.
func (f *fiberContext) BindJSON(obj any) error {
	return f.Ctx.BodyParser(obj)
}

// Set implements the Context interface.
func (f *fiberContext) Set(key string, val any) {
	f.Ctx.Locals(key, val)
}

// RequestContext implements the Context interface.
func (f *fiberContext) RequestContext() context.Context {
	return f.Ctx.UserContext()
}

// WrapHandlerFiber wraps a handler function into a Fiber handler.
func WrapHandlerFiber[T any, R any](handle WrapHandlerFunc[T, R]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return WrapHandler(handle)(&fiberContext{Ctx: c})
	}
}
