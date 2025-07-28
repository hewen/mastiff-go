// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/panjf2000/gnet/v2"
)

// GnetHandler is a handler that provides a unified socket abstraction over gnet.
type GnetHandler struct {
	logger logger.Logger
	conf   *serverconf.SocketConfig
	event  gnet.EventHandler
	engine gnet.Engine
}

// OnBoot is called once when the engine starts.
func (h *GnetHandler) OnBoot(e gnet.Engine) gnet.Action {
	if h.event == nil {
		return gnet.Close
	}

	h.engine = e
	return h.event.OnBoot(e)
}

// OnShutdown is called when the engine is shutting down.
func (h *GnetHandler) OnShutdown(e gnet.Engine) {
	if h.event == nil {
		return
	}
	h.event.OnShutdown(e)
}

// OnOpen is called when a new connection is opened.
func (h *GnetHandler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	if h.event == nil {
		return nil, gnet.Close
	}

	return h.event.OnOpen(c)
}

// OnClose is called when a connection is closed.
func (h *GnetHandler) OnClose(c gnet.Conn, err error) gnet.Action {
	if h.event == nil {
		return gnet.Close
	}

	return h.event.OnClose(c, err)
}

// OnTick is called periodically by the event loop.
func (h *GnetHandler) OnTick() (time.Duration, gnet.Action) {
	if h.event == nil {
		return h.conf.TickInterval, gnet.None
	}

	return h.event.OnTick()
}

// OnTraffic is triggered when a complete message is received.
func (h *GnetHandler) OnTraffic(c gnet.Conn) gnet.Action {
	if h.event == nil {
		return gnet.Close
	}

	return h.event.OnTraffic(c)
}

// Name returns the name of the GnetHandler.
func (h *GnetHandler) Name() string {
	return "gnet"
}

// Start starts the GnetHandler.
func (h *GnetHandler) Start() error {
	return gnet.Run(h, h.conf.Addr, gnet.WithOptions(h.conf.GnetOptions))
}

// Stop stops the GnetHandler.
func (h *GnetHandler) Stop() error {
	return h.engine.Stop(context.TODO())
}

// NewGnetHandler creates a new GnetHandler.
func NewGnetHandler(conf *serverconf.SocketConfig, event gnet.EventHandler) (*GnetHandler, error) {
	if conf == nil {
		return nil, ErrEmptySocketConf
	}

	conf.SetDefault()

	l := logger.NewLogger()
	conf.GnetOptions.Logger = l

	return &GnetHandler{
		logger: l,
		conf:   conf,
		event:  event,
	}, nil
}
