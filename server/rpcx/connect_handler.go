// Package rpcx provides a unified RPC abstraction over gRPC and Connect.
package rpcx

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ConnectHandler is a handler that provides a unified RPC abstraction over Connect.
type ConnectHandler struct {
	server *http.Server
	ln     net.Listener
	addr   string
}

// ConnectHandlerBuilder builds a Connect handler.
type ConnectHandlerBuilder struct {
	RegisterMux func(mux *http.ServeMux)
	Conf        *serverconf.RPCConfig
}

// BuildRPC builds a Connect handler.
func (b *ConnectHandlerBuilder) BuildRPC() (RPCHandler, error) {
	if b.Conf == nil {
		return nil, ErrEmptyRPCConf
	}
	if b.RegisterMux == nil {
		return nil, fmt.Errorf("connect: register mux is nil")
	}

	mux := http.NewServeMux()
	b.RegisterMux(mux)

	handler := h2c.NewHandler(mux, &http2.Server{})

	if b.Conf.Timeout == 0 {
		b.Conf.Timeout = RPCTimeoutDefault
	}

	srv := http.Server{
		Addr:         b.Conf.Addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(b.Conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(b.Conf.Timeout) * time.Second,
	}

	ln, err := net.Listen("tcp", b.Conf.Addr)
	if err != nil {
		return nil, err
	}

	return &ConnectHandler{
		server: &srv,
		ln:     ln,
		addr:   b.Conf.Addr,
	}, nil
}

// Start starts the Connect handler.
func (h *ConnectHandler) Start() error {
	return h.server.Serve(h.ln)
}

// Stop stops the Connect handler.
func (h *ConnectHandler) Stop() error {
	return h.server.Close()
}

// Name returns the name of the Connect handler.
func (h *ConnectHandler) Name() string {
	return fmt.Sprintf("rpc connect(%s)", h.addr)
}
