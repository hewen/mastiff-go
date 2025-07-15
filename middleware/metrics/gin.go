// Package metrics provides a middleware for recording HTTP request metrics in Gin framework.
package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware is a middleware for recording HTTP request metrics in Gin framework.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		HTTPDuration.WithLabelValues(
			c.Request.Method,
			path,
			httpStatusCodeGroup(c.Writer.Status()),
		).Observe(time.Since(start).Seconds())
	}
}

// httpStatusCodeGroup groups HTTP status codes into categories.
func httpStatusCodeGroup(status int) string {
	return string(rune(status / 100 * 100)) // e.g., 200 -> "200"
}
