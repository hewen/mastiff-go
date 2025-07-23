// Package handler provides a unified HTTP abstraction over Gin.
package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
)

// GinHandler is a handler that provides a unified HTTP abstraction over Gin.
type GinHandler struct {
	RouterGroup
	name      string
	addr      string
	ginEngine *gin.Engine
	server    http.Server
}

// Start starts the GinHandler.
func (g *GinHandler) Start() error {
	return g.server.ListenAndServe()
}

// Stop stops the GinHandler.
func (g *GinHandler) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return g.server.Shutdown(ctx)
}

// Name returns the name of the GinHandler.
func (g *GinHandler) Name() string {
	return fmt.Sprintf("http %s server(%s)", g.name, g.addr)
}

// Use adds middleware to the router.
func (g *GinHandler) Use(handler ...HTTPHandlerFunc) Router {
	return newGinRouter(g.ginEngine.Use(AsGinHandler(handler...)...))
}

// Test sends a request to the GinHandler and returns the response.
func (g *GinHandler) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	w := httptest.NewRecorder()

	if len(msTimeout) > 0 && msTimeout[0] > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(msTimeout[0])*time.Millisecond)
		defer cancel()
		req = req.WithContext(ctx)
	}

	g.ginEngine.ServeHTTP(w, req)

	resp := w.Result()
	resp.Body = io.NopCloser(bytes.NewBuffer(w.Body.Bytes()))
	return resp, nil
}

// NewGinHandler creates a new GinHandler.
func NewGinHandler(conf *serverconf.HTTPConfig) (HTTPHandler, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	gin.SetMode(conf.Mode)
	r := gin.New()

	return &GinHandler{
		RouterGroup: newGinRouterGroup(&r.RouterGroup),
		ginEngine:   r,
		name:        "gin",
		addr:        conf.Addr,
		server: http.Server{
			Addr:         conf.Addr,
			Handler:      r,
			ReadTimeout:  toDuration(conf.ReadTimeout),
			WriteTimeout: toDuration(conf.WriteTimeout),
			IdleTimeout:  toDuration(conf.IdleTimeout),
		},
	}, nil
}

// GinRouterGroup implements the RouterGroup interface for Gin.
type GinRouterGroup struct {
	Router
	r *gin.RouterGroup
}

// Group creates a new router group with the given relative path and handlers.
func (group *GinRouterGroup) Group(relativePath string, handlers ...HTTPHandlerFunc) RouterGroup {
	return newGinRouterGroup(group.r.Group(relativePath, AsGinHandler(handlers...)...))
}

// GinRouter implements the Router interface for Gin.
type GinRouter struct {
	r gin.IRoutes
}

// Use adds middleware to the router.
func (g *GinRouter) Use(handlers ...HTTPHandlerFunc) Router {
	g.r.Use(AsGinHandler(handlers...)...)
	return g
}

// Handle adds a route with the given method and path.
func (g *GinRouter) Handle(method, path string, handlers ...HTTPHandlerFunc) Router {
	g.r.Handle(method, path, AsGinHandler(handlers...)...)
	return g
}

// Any adds a route that matches all HTTP methods.
func (g *GinRouter) Any(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.Any(path, AsGinHandler(handlers...)...)
	return g
}

// Get adds a route that matches GET requests.
func (g *GinRouter) Get(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.GET(path, AsGinHandler(handlers...)...)
	return g
}

// Post adds a route that matches POST requests.
func (g *GinRouter) Post(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.POST(path, AsGinHandler(handlers...)...)
	return g
}

// Delete adds a route that matches DELETE requests.
func (g *GinRouter) Delete(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.DELETE(path, AsGinHandler(handlers...)...)
	return g
}

// Patch adds a route that matches PATCH requests.
func (g *GinRouter) Patch(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.PATCH(path, AsGinHandler(handlers...)...)
	return g
}

// Put adds a route that matches PUT requests.
func (g *GinRouter) Put(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.PUT(path, AsGinHandler(handlers...)...)
	return g
}

// Options adds a route that matches OPTIONS requests.
func (g *GinRouter) Options(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.OPTIONS(path, AsGinHandler(handlers...)...)
	return g
}

// Head adds a route that matches HEAD requests.
func (g *GinRouter) Head(path string, handlers ...HTTPHandlerFunc) Router {
	g.r.HEAD(path, AsGinHandler(handlers...)...)
	return g
}

// Match adds a route that matches the given HTTP methods.
func (g *GinRouter) Match(methods []string, path string, handlers ...HTTPHandlerFunc) Router {
	g.r.Match(methods, path, AsGinHandler(handlers...)...)
	return g
}

// newGinRouterGroup creates a new RouterGroup by gin.RouterGroup.
func newGinRouterGroup(r *gin.RouterGroup) RouterGroup {
	return &GinRouterGroup{
		Router: newGinRouter(r),
		r:      r,
	}
}

// newGinRouter creates a new Router by gin.IRoutes.
func newGinRouter(r gin.IRoutes) Router {
	return &GinRouter{r: r}
}
