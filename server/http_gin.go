// Package server provides a Gin server implementation.
package server

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/middleware/recovery"
)

// NewGinAPIHandler initializes a new Gin API handler with the provided route initialization function.
func NewGinAPIHandler(conf *HTTPConf, initRoute func(r *gin.Engine)) http.Handler {
	gin.SetMode(conf.Mode)
	r := gin.New()
	r.Use(recovery.GinRecoverHandler())
	r.Use(logging.GinLoggingHandler())

	if conf.PprofEnabled {
		pprof.Register(r)
	}

	initRoute(r)
	return r
}
