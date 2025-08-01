// Package circuitbreaker provides a circuit breaker middleware for Fiber.
package circuitbreaker

import (
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a middleware for circuit breaker in httpx framework.
func HttpxMiddleware(mgr *Manager) func(c unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		key := c.FullPath()
		breaker := mgr.Get(key)

		_, err := breaker.Execute(func() (any, error) {
			err := c.Next()
			return nil, err
		})

		return err
	}
}
