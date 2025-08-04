// Package recovery provides a middleware for recovering from panics in Fiber framework.
package recovery

import (
	"runtime/debug"
	"strings"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// HttpxMiddleware is a middleware for recovering from panics in framework.
func HttpxMiddleware() func(c unicontext.UniversalContext) error {
	return func(c unicontext.UniversalContext) error {
		defer func() {
			if r := recover(); r != nil {
				l := logger.NewLoggerWithContext(unicontext.ContextFrom(c))
				l.Errorf("panic: %v $%s", r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
			}
		}()
		return c.Next()
	}
}
