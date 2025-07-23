// Package handler provides a unified HTTP abstraction over Gin and Fiber.
package handler

import (
	"net/http/pprof"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHandler creates a new HTTP handler.
func NewHandler(conf *serverconf.HTTPConfig, opts ...ServerOption) (HTTPHandler, error) {
	if conf == nil {
		return nil, ErrEmptyHTTPConf
	}

	var h HTTPHandler
	var err error
	switch conf.FrameworkType {
	case serverconf.FrameworkGin:
		h, err = NewGinHandler(conf)
	case serverconf.FrameworkFiber:
		h, err = NewFiberHandler(conf)
	default:
		panic("unknown framework type")
	}
	if err != nil {
		return nil, err
	}

	for i := range opts {
		opts[i](h)
	}

	return h, nil
}

// ServerOption is a function that configures a HTTP server.
type ServerOption func(HTTPHandler)

// WithMetrics adds a metrics handler to the server.
func WithMetrics() ServerOption {
	return func(h HTTPHandler) {
		h.Get("/metrics", FromHTTPHandler(promhttp.Handler()))
	}
}

// WithPprof adds a pprof handler to the server.
func WithPprof() ServerOption {
	return func(h HTTPHandler) {
		h.Get("/debug/pprof/", FromHTTPHandlerFunc(pprof.Index))
		h.Get("/debug/pprof/cmdline", FromHTTPHandlerFunc(pprof.Cmdline))
		h.Get("/debug/pprof/profile", FromHTTPHandlerFunc(pprof.Profile))
		h.Get("/debug/pprof/symbol", FromHTTPHandlerFunc(pprof.Symbol))
		h.Get("/debug/pprof/trace", FromHTTPHandlerFunc(pprof.Trace))
	}
}
