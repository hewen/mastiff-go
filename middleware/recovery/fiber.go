// Package recovery provides a middleware for recovering from panics in Fiber framework.
package recovery

import (
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
)

// FiberMiddleware is a middleware for recovering from panics in Fiber framework.
func FiberMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				l := logger.NewLoggerWithContext(contextkeys.ContextFrom(c))
				l.Errorf("panic: %v $%s", r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
			}
		}()
		return c.Next()
	}
}
