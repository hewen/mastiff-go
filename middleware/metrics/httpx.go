// Package metrics provides a middleware for recording HTTP request metrics in Fiber framework.
package metrics

import (
	"time"

	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a middleware for recording HTTP request metrics in httpx framework.
func HttpxMiddleware() func(c unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		start := time.Now()

		err := c.Next()

		path := c.FullPath()
		HTTPDuration.WithLabelValues(
			c.Method(),
			path,
			httpStatusCodeGroup(c.StatusCode()),
		).Observe(time.Since(start).Seconds())
		return err
	}
}

// httpStatusCodeGroup groups HTTP status codes into categories.
func httpStatusCodeGroup(status int) string {
	return string(rune(status / 100 * 100)) // e.g., 200 -> "200"
}
