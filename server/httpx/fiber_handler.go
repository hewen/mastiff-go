// Package httpx provides a unified HTTP abstraction over Gin and Fiber.
package httpx

import (
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// FiberHandlerBuilder builds a Fiber HTTP handler.
type FiberHandlerBuilder struct {
	InitRoute        func(app *fiber.App)
	ExtraMiddlewares []func(*fiber.Ctx) error
	Conf             serverconf.HTTPConfig
}

// BuildHandler builds a Fiber HTTP handler.
func (f *FiberHandlerBuilder) BuildHandler() http.Handler {
	app := fiber.New()

	for _, mw := range middleware.LoadFiberMiddlewares(f.Conf.Middlewares) {
		app.Use(mw)
	}

	for _, mw := range f.ExtraMiddlewares {
		app.Use(mw)
	}

	if f.Conf.PprofEnabled {
		app.Use(pprof.New())
	}
	if f.Conf.Middlewares.EnableMetrics != nil && *f.Conf.Middlewares.EnableMetrics {
		app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	}

	f.InitRoute(app)

	return adaptor.FiberApp(app)
}
