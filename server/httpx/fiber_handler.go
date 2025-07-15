// Package httpx provides a unified HTTP abstraction over Gin and Fiber.
package httpx

import (
	"fmt"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// FiberHandlerBuilder builds a Fiber HTTP handler.
type FiberHandlerBuilder struct {
	Conf             *serverconf.HTTPConfig
	InitRoute        func(app *fiber.App)
	ExtraMiddlewares []func(*fiber.Ctx) error
}

// BuildHandler builds a Fiber HTTP handler.
func (f *FiberHandlerBuilder) BuildHandler() (HTTPHandler, error) {
	if f.Conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(f.Conf.ReadTimeout),
		WriteTimeout: time.Duration(f.Conf.WriteTimeout),
	})

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

	return &FiberHandler{
		app:  app,
		addr: f.Conf.Addr,
		name: "fiber",
	}, nil
}

// FiberHandler is a handler that provides a unified HTTP abstraction over Fiber.
type FiberHandler struct {
	app  *fiber.App
	addr string
	name string
}

// Start starts the FiberHandler.
func (s *FiberHandler) Start() error {
	return s.app.Listen(s.addr)
}

// Stop stops the FiberHandler.
func (s *FiberHandler) Stop() error {
	return s.app.Shutdown()
}

// Name returns the name of the FiberHandler.
func (s *FiberHandler) Name() string {
	return fmt.Sprintf("http %s server(%s)", s.name, s.addr)
}
