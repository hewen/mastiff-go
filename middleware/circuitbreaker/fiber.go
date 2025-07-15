// Package circuitbreaker provides a circuit breaker middleware for Fiber.
package circuitbreaker

import (
	"github.com/gofiber/fiber/v2"
)

// FiberMiddleware is a middleware for circuit breaker in Fiber framework.
func FiberMiddleware(mgr *Manager) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		key := c.Path()
		breaker := mgr.Get(key)

		_, err := breaker.Execute(func() (any, error) {
			err := c.Next()
			return nil, err
		})

		return err
	}
}
