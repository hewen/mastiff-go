// Package httpx provides a unified HTTP abstraction over Gin and Fiber.
package httpx

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GinHandlerBuilder builds a Gin HTTP handler.
type GinHandlerBuilder struct {
	InitRoute        func(r *gin.Engine)
	ExtraMiddlewares []gin.HandlerFunc
	Conf             serverconf.HTTPConfig
}

// BuildHandler builds a Gin HTTP handler.
func (g *GinHandlerBuilder) BuildHandler() http.Handler {
	gin.SetMode(g.Conf.Mode)
	r := gin.New()

	for _, mw := range middleware.LoadGinMiddlewares(g.Conf.Middlewares) {
		r.Use(mw)
	}

	for _, mw := range g.ExtraMiddlewares {
		r.Use(mw)
	}

	if g.Conf.PprofEnabled {
		pprof.Register(r)
	}
	if g.Conf.Middlewares.EnableMetrics != nil && *g.Conf.Middlewares.EnableMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	g.InitRoute(r)
	return r
}
