// Package metrics provides a middleware for recording HTTP request metrics in Fiber framework.
package metrics

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// FiberMiddleware is a middleware for recording HTTP request metrics in Fiber framework.
func FiberMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		path := c.Path()
		HTTPDuration.WithLabelValues(
			c.Method(),
			path,
			httpStatusCodeGroup(c.Response().StatusCode()),
		).Observe(time.Since(start).Seconds())
		return err
	}
}
