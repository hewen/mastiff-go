// Package server gin server implementation
package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
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
			limited := io.LimitReader(c.Request.Body, 1<<20) // max 1MB
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

		l := logger.NewLoggerWithTraceID(traceID)
		isStatic := isStaticResource(c.Request.URL.Path, respContentType)
		if isStatic {
			requestBody = ""
			responseBody = ""
		}

		var err error
		if c.Errors != nil {
			err = c.Errors.Last().Err
		}

		LogRequest(
			l,
			c.Writer.Status(),
			time.Since(start),
			realip.FromRequest(c.Request),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			c.Request.UserAgent(),
			requestBody,
			responseBody,
			err,
		)
	}
}

// isTextContent checks if the content type is text-like.
func isTextContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "application/xml") ||
		strings.Contains(ct, "application/x-www-form-urlencoded") ||
		strings.Contains(ct, "text/") ||
		strings.Contains(ct, "application/javascript")
}

// isStaticResource checks if the path or content-type indicates a static resource.
func isStaticResource(path string, contentType string) bool {
	exts := []string{
		".js", ".css", ".html", ".ico", ".svg", ".ttf", ".woff", ".woff2",
		".eot", ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".map", ".json",
	}
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "image/") ||
		strings.Contains(ct, "text/css") ||
		strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "application/javascript") ||
		strings.Contains(ct, "font/")
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
func NewGinAPIHandler(conf *HTTPConf, initRoute func(r *gin.Engine)) http.Handler {
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
