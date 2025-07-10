// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRateLimitHandler creates a Gin middleware with rate limiter.
func GinRateLimitHandler(mgr *LimiterManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		cfg := mgr.config.PerRoute[route]
		if cfg == nil {
			cfg = mgr.config.Default
		}
		if cfg == nil {
			c.Next()
			return
		}
		key := mgr.getKeyFromGin(c, cfg)
		limiter := mgr.getOrCreateLimiter(key, cfg)
		if err := limiter.AllowOrWait(c.Request.Context()); err != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
