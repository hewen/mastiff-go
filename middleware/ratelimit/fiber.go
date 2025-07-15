// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/internal/contextkeys"
)

// FiberMiddleware creates a Fiber middleware with rate limiter.
func FiberMiddleware(mgr *LimiterManager) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		route := c.Path()
		cfg := mgr.config.PerRoute[route]
		if cfg == nil {
			cfg = mgr.config.Default
		}
		if cfg == nil {
			return c.Next()
		}

		key := mgr.getKeyFromFiber(c, cfg)
		limiter := mgr.getOrCreateLimiter(key, cfg)
		ctx := contextkeys.ContextFrom(c)
		if err := limiter.AllowOrWait(ctx); err != nil {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}
