// Package handler provides a unified RPC abstraction over gRPC and Connect.
package handler

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcHandler is a handler that provides a unified RPC abstraction over gRPC.
type GrpcHandler struct {
	s    *grpc.Server
	ln   net.Listener
	addr string
}

// NewGrpcHandler builds a gRPC handler.
func NewGrpcHandler(
	conf *serverconf.RPCConfig,
	registerFunc func(*grpc.Server),
	extraInterceptors ...grpc.UnaryServerInterceptor,
) (RPCHandler, error) {
	if conf == nil {
		return nil, ErrEmptyRPCConf
	}

	interceptors := middleware.LoadGRPCMiddlewares(conf.Middlewares)
	interceptors = append(interceptors, extraInterceptors...)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
	}

	s := grpc.NewServer(opts...)
	registerFunc(s)

	if conf.Reflection {
		reflection.Register(s)
	}

	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		return nil, err
	}

	return &GrpcHandler{
		s:    s,
		ln:   ln,
		addr: conf.Addr,
	}, nil
}

// Start starts the gRPC handler.
func (h *GrpcHandler) Start() error {
	return h.s.Serve(h.ln)
}

// Stop stops the gRPC handler.
func (h *GrpcHandler) Stop() error {
	h.s.GracefulStop()
	return h.ln.Close()
}

// Name returns the name of the gRPC handler.
func (h *GrpcHandler) Name() string {
	return fmt.Sprintf("rpc grpc(%s)", h.addr)
}
