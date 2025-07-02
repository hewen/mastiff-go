// Package server gin server implementation
package server

import (
	"bytes"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"github.com/tomasen/realip"
)

// GinLoggerHandler is a middleware for logging HTTP requests in Gin framework.
func GinLoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		begin := time.Now()
		var bodyBytes []byte

		if c.Request.Header.Get("Content-type") == "application/json" {
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		traceID := logger.NewTraceID()
		c.Set(string(logger.LoggerTraceKey), traceID)

		c.Next()

		bodyStr := strings.ReplaceAll(string(bodyBytes), "\n", "")

		l := logger.NewLoggerWithTraceID(traceID)
		l.Infof("%3d | %10s | %15s | %-7s | %s | %s | %s |  %s | %s| %s | %v",
			c.Writer.Status(),
			util.FormatDuration(time.Since(begin)),
			realip.FromRequest(c.Request),
			c.Request.Method,
			c.Request.Header.Get("Content-type"),
			c.Request.Host,
			c.Request.URL,
			c.Request.Proto,
			c.Request.UserAgent(),
			bodyStr,
			c.Errors,
		)
	}
}

// GinRecoverHandler is a middleware for recovering from panics in Gin framework.
func GinRecoverHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				l := logger.NewLoggerWithGinContext(c)
				l.Panicf("%s | %s | %s | $%s", realip.FromRequest(c.Request), c.Request.UserAgent(), r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
			}
		}()
		c.Next()
	}
}

// NewGinAPIHandler initializes a new Gin API handler with the provided route initialization function.
func NewGinAPIHandler(initRoute func(r *gin.Engine)) http.Handler {
	r := gin.New()
	r.Use(GinRecoverHandler())
	r.Use(GinLoggerHandler())

	initRoute(r)
	return r
}
