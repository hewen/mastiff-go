// Package logging provides a middleware for logging HTTP requests in Gin framework.
package logging

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
	"github.com/tomasen/realip"
)

// GinMiddleware is a middleware for logging HTTP requests in Gin framework.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		ctx := logger.NewOutgoingContextWithIncomingContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		req, _ := c.Get("req")
		resp, _ := c.Get("resp")

		var err error
		if c.Errors != nil {
			err = c.Errors.Last()
		}

		l := logger.NewLoggerWithContext(ctx)
		logger.LogRequest(
			l,
			c.Writer.Status(),
			time.Since(start),
			realip.FromRequest(c.Request),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			c.Request.UserAgent(),
			req,
			resp,
			err,
		)
	}
}
