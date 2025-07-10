// Package recovery provides a middleware for recovering from panics in Gin framework.
package recovery

import (
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
)

// GinRecoverHandler is a middleware for recovering from panics in Gin framework.
func GinRecoverHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				l := logger.NewLoggerWithContext(c.Request.Context())
				l.Errorf("panic: %v $%s", r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
			}
		}()
		c.Next()
	}
}
