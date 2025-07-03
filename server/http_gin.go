// Package server gin server implementation
package server

import (
	"bytes"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"github.com/tomasen/realip"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // capture response body
	return w.ResponseWriter.Write(b)
}

// GinLoggerHandler is a middleware for logging HTTP requests in Gin framework.
func GinLoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var requestBody, responseBody string
		var reqBodyBytes []byte

		reqContentType := c.Request.Header.Get("Content-Type")
		if c.Request.Body != nil {
			limited := io.LimitReader(c.Request.Body, 1<<20) // 最大 1MB
			reqBodyBytes, _ = io.ReadAll(limited)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		if isTextContent(reqContentType) {
			requestBody = strings.ReplaceAll(string(reqBodyBytes), "\n", "")
		} else if len(reqBodyBytes) > 0 {
			requestBody = "[binary]"
		}

		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = blw

		traceID := logger.NewTraceID()
		c.Set(string(logger.LoggerTraceKey), traceID)

		c.Next()

		respContentType := c.Writer.Header().Get("Content-Type")
		if isTextContent(respContentType) {
			responseBody = strings.TrimSpace(blw.body.String())
		} else if blw.body.Len() > 0 {
			responseBody = "[binary]"
		}

		log := logger.NewLoggerWithTraceID(traceID)
		log.Infof(
			"%3d | %10s | %15s | %-7s | %s | %s | %s | %s | UA: %s | req: %s | resp: %s",
			c.Writer.Status(),
			util.FormatDuration(time.Since(start)),
			realip.FromRequest(c.Request),
			c.Request.Method,
			reqContentType,
			c.Request.Host,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Request.UserAgent(),
			requestBody,
			responseBody,
		)

		if c.Errors != nil {
			log.Errorf("%v", c.Errors)
		}
	}
}

func isTextContent(contentType string) bool {
	ct := strings.ToLower(contentType)

	return strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "application/xml") ||
		strings.Contains(ct, "application/x-www-form-urlencoded") ||
		strings.Contains(ct, "text/") ||
		strings.Contains(ct, "application/javascript")
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
func NewGinAPIHandler(conf *HTTPConfig, initRoute func(r *gin.Engine)) http.Handler {
	gin.SetMode(conf.Mode)
	r := gin.New()
	r.Use(GinRecoverHandler())
	r.Use(GinLoggerHandler())

	if conf.PprofEnabled {
		pprof.Register(r)
	}

	initRoute(r)
	return r
}
