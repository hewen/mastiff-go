// Package logging provides a middleware for logging HTTP requests in Gin framework.
package logging

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/tomasen/realip"
)

// GinLoggingHandler is a middleware for logging HTTP requests in Gin framework.
func GinLoggingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := logger.NewTraceID()

		c.Set(string(contextkeys.LoggerTraceIDKey), traceID)

		c.Next()

		req, _ := c.Get("req")
		resp, _ := c.Get("resp")

		l := logger.NewLoggerWithGinContext(c)
		logger.LogRequest(
			l,
			c.Writer.Status(),
			time.Since(start),
			realip.FromRequest(c.Request),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			c.Request.UserAgent(),
			req,
			resp,
			nil,
		)
	}
}
