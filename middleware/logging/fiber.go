// Package logging provides a middleware for logging HTTP requests in Gin framework.
package logging

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
)

// FiberMiddleware is a middleware for logging HTTP requests in Fiber framework.
func FiberMiddleware() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		ctx := contextkeys.ContextFrom(c)
		ctx = logger.NewOutgoingContextWithIncomingContext(ctx)
		contextkeys.InjectContext(ctx, c)

		err := c.Next()

		req := c.Locals("req")
		resp := c.Locals("resp")

		l := logger.NewLoggerWithContext(ctx)

		logger.LogRequest(
			l,
			c.Response().StatusCode(),
			time.Since(start),
			c.Context().RemoteIP().String(),
			fmt.Sprintf("%s %s", c.Method(), c.Path()),
			string(c.Request().Header.UserAgent()),
			req,
			resp,
			err,
		)

		return err
	}
}
