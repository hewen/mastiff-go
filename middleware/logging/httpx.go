// Package logging provides a middleware for logging HTTP requests in Gin framework.
package logging

import (
	"fmt"
	"time"

	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a middleware for logging HTTP requests in Fiber framework.
func HttpxMiddleware() func(unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		start := time.Now()

		ctx := contextkeys.ContextFrom(c)
		ctx = logger.NewOutgoingContextWithIncomingContext(ctx)
		contextkeys.InjectContext(ctx, c)

		err := c.Next()

		req, _ := c.Get("req")
		resp, _ := c.Get("resp")

		l := logger.NewLoggerWithContext(ctx)

		logger.LogRequest(
			l,
			c.StatusCode(),
			time.Since(start),
			c.ClientIP(),
			fmt.Sprintf("%s %s", c.Method(), c.Path()),
			c.Request().UserAgent(),
			req,
			resp,
			err,
		)

		return err
	}
}
