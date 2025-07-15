// Package rpcx provides a unified RPC abstraction over gRPC and Connect.
package rpcx

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

// GrpcHandlerBuilder builds a gRPC handler.
type GrpcHandlerBuilder struct {
	Conf              *serverconf.RPCConfig
	RegisterFunc      func(*grpc.Server)
	ExtraInterceptors []grpc.UnaryServerInterceptor
}

// BuildRPC builds a gRPC handler.
func (b *GrpcHandlerBuilder) BuildRPC() (RPCHandler, error) {
	if b.Conf == nil {
		return nil, ErrEmptyRPCConf
	}

	interceptors := middleware.LoadGRPCMiddlewares(b.Conf.Middlewares)
	interceptors = append(interceptors, b.ExtraInterceptors...)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
	}

	s := grpc.NewServer(opts...)
	b.RegisterFunc(s)

	if b.Conf.Reflection {
		reflection.Register(s)
	}

	ln, err := net.Listen("tcp", b.Conf.Addr)
	if err != nil {
		return nil, err
	}

	return &GrpcHandler{
		s:    s,
		ln:   ln,
		addr: b.Conf.Addr,
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
