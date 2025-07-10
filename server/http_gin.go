// Package server provides a Gin server implementation.
package server

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewGinAPIHandler initializes a new Gin API handler with the provided route initialization function.
func NewGinAPIHandler(conf *HTTPConf, initRoute func(r *gin.Engine), extraMiddlewares ...gin.HandlerFunc) (http.Handler, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	gin.SetMode(conf.Mode)
	r := gin.New()

	// Load and apply framework middlewares from config
	mws := middleware.LoadGinMiddlewares(conf.Middlewares)
	for _, mw := range mws {
		r.Use(mw)
	}

	// Apply additional user-provided middlewares
	for _, mw := range extraMiddlewares {
		r.Use(mw)
	}

	// Register pprof if enabled
	if conf.PprofEnabled {
		pprof.Register(r)
	}

	// Register metrics endpoint if enabled
	if conf.Middlewares.EnableMetrics != nil && *conf.Middlewares.EnableMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// Route setup
	initRoute(r)
	return r, nil
}
