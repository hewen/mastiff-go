// Package handler provides a unified RPC abstraction over gRPC and Connect.
package handler

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

// NewConnectHandler builds a Connect handler.
func NewConnectHandler(conf *serverconf.RPCConfig, registerMux func(mux *http.ServeMux)) (RPCHandler, error) {
	if conf == nil {
		return nil, ErrEmptyRPCConf
	}
	if registerMux == nil {
		return nil, fmt.Errorf("connect: register mux is nil")
	}

	mux := http.NewServeMux()
	registerMux(mux)

	handler := h2c.NewHandler(mux, &http2.Server{})

	if conf.Timeout == 0 {
		conf.Timeout = RPCTimeoutDefault
	}

	srv := http.Server{
		Addr:         conf.Addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Timeout) * time.Second,
	}

	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		return nil, err
	}

	return &ConnectHandler{
		server: &srv,
		ln:     ln,
		addr:   conf.Addr,
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
