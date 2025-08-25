// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"net/http"

	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware creates a middleware with rate limiter.
func HttpxMiddleware(mgr *LimiterManager) func(c unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		route := c.FullPath()
		cfg := mgr.config.PerRoute[route]
		if cfg == nil {
			cfg = mgr.config.Default
		}
		if cfg == nil {
			return c.Next()
		}

		key := mgr.getKeyFromHttpx(c, cfg)
		limiter := mgr.getOrCreateLimiter(key, cfg)
		ctx := unicontext.ContextFrom(c)
		if err := limiter.AllowOrWait(ctx); err != nil {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}
