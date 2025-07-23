// Package handler provides a unified HTTP abstraction over Fiber.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/serverconf"
)

// FiberHandler is a handler that provides a unified HTTP abstraction over Fiber.
type FiberHandler struct {
	RouterGroup
	app  *fiber.App
	addr string
	name string
}

// Start starts the FiberHandler.
func (f *FiberHandler) Start() error {
	return f.app.Listen(f.addr)
}

// Stop stops the FiberHandler.
func (f *FiberHandler) Stop() error {
	return f.app.Shutdown()
}

// Name returns the name of the FiberHandler.
func (f *FiberHandler) Name() string {
	return fmt.Sprintf("http %s server(%s)", f.name, f.addr)
}

// Test sends a request to the FiberHandler and returns the response.
func (f *FiberHandler) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return f.app.Test(req, msTimeout...)
}

// Use adds middleware to the router.
func (f *FiberHandler) Use(handlers ...UniversalHandlerFunc) Router {
	args := make([]any, len(handlers))
	h := AsFiberHandler(handlers...)
	for i := range h {
		args[i] = h[i]
	}
	return newFiberRouterGroup(f.app.Use(args...))
}

// getFiberConfig returns the Fiber configuration.
func getFiberConfig(mode string) fiber.Config {
	if mode == "release" {
		return fiber.Config{
			Prefork:               true,
			CaseSensitive:         true,
			StrictRouting:         true,
			EnablePrintRoutes:     false,
			DisableStartupMessage: true,
		}
	}

	return fiber.Config{
		Prefork:               false,
		CaseSensitive:         true,
		StrictRouting:         false,
		EnablePrintRoutes:     false,
		DisableStartupMessage: true,
	}
}

// NewFiberHandler creates a new FiberHandler.
func NewFiberHandler(conf *serverconf.HTTPConfig) (UniversalHandler, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	fiberConfig := getFiberConfig(conf.Mode)
	fiberConfig.ReadTimeout = toDuration(conf.ReadTimeout)
	fiberConfig.WriteTimeout = toDuration(conf.WriteTimeout)
	fiberConfig.IdleTimeout = toDuration(conf.IdleTimeout)

	app := fiber.New(fiberConfig)

	return &FiberHandler{
		RouterGroup: newFiberRouterGroup(app),
		app:         app,
		addr:        conf.Addr,
		name:        "fiber",
	}, nil
}

// FiberRouterGroup implements the RouterGroup interface for Fiber.
type FiberRouterGroup struct {
	Router
	r fiber.Router
}

// Group creates a new router group with the given relative path and handlers.
func (group *FiberRouterGroup) Group(relativePath string, handlers ...UniversalHandlerFunc) RouterGroup {
	return newFiberRouterGroup(group.r.Group(relativePath, AsFiberHandler(handlers...)...))
}

// FiberRouter implements the Router interface for Fiber.
type FiberRouter struct {
	Router
	r fiber.Router
}

// Use adds middleware to the router.
func (f *FiberRouter) Use(handlers ...UniversalHandlerFunc) Router {
	args := make([]any, len(handlers))
	h := AsFiberHandler(handlers...)
	for i := range h {
		args[i] = h[i]
	}
	return newFiberRouterGroup(f.r.Use(args...))
}

// Handle adds a route with the given method and path.
func (f *FiberRouter) Handle(method, path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Add(method, path, AsFiberHandler(handlers...)...)
	return f
}

// Any adds a route that matches all HTTP methods.
func (f *FiberRouter) Any(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.All(path, AsFiberHandler(handlers...)...)
	return f
}

// Get adds a route that matches GET requests.
func (f *FiberRouter) Get(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Get(path, AsFiberHandler(handlers...)...)
	return f
}

// Post adds a route that matches POST requests.
func (f *FiberRouter) Post(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Post(path, AsFiberHandler(handlers...)...)
	return f
}

// Delete adds a route that matches DELETE requests.
func (f *FiberRouter) Delete(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Delete(path, AsFiberHandler(handlers...)...)
	return f
}

// Patch adds a route that matches PATCH requests.
func (f *FiberRouter) Patch(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Patch(path, AsFiberHandler(handlers...)...)
	return f
}

// Put adds a route that matches PUT requests.
func (f *FiberRouter) Put(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Put(path, AsFiberHandler(handlers...)...)
	return f
}

// Options adds a route that matches OPTIONS requests.
func (f *FiberRouter) Options(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Options(path, AsFiberHandler(handlers...)...)
	return f
}

// Head adds a route that matches HEAD requests.
func (f *FiberRouter) Head(path string, handlers ...UniversalHandlerFunc) Router {
	f.r.Head(path, AsFiberHandler(handlers...)...)
	return f
}

// Match adds a route that matches the given HTTP methods.
func (f *FiberRouter) Match(methods []string, path string, handlers ...UniversalHandlerFunc) Router {
	for _, method := range methods {
		f.r.Add(method, path, AsFiberHandler(handlers...)...)
	}
	return f
}

// newFiberRouterGroup creates a new RouterGroup by fiber.Router.
func newFiberRouterGroup(r fiber.Router) RouterGroup {
	return &FiberRouterGroup{
		Router: newFiberRouter(r),
		r:      r,
	}
}

// newFiberRouter creates a new Router by fiber.Router.
func newFiberRouter(r fiber.Router) Router {
	return &FiberRouter{r: r}
}
