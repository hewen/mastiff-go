// Package handler provides a unified RPC abstraction over gRPC and Connect.
package handler

import (
	"fmt"
	"net/http"

	"github.com/hewen/mastiff-go/config/serverconf"
	"google.golang.org/grpc"
)

// RPCBuildParams contains the parameters needed to build a RPC handler.
type RPCBuildParams struct {
	GrpcRegisterFunc      func(*grpc.Server)
	ConnectRegisterMux    func(*http.ServeMux)
	ExtraGrpcInterceptors []grpc.UnaryServerInterceptor
}

// NewHandler creates a handler from registered builders for different RPC frameworks.
func NewHandler(conf *serverconf.RPCConfig, params RPCBuildParams) (RPCHandler, error) {
	if conf == nil {
		return nil, ErrEmptyRPCConf
	}

	switch conf.FrameworkType {
	case serverconf.FrameworkGrpc:
		if params.GrpcRegisterFunc == nil {
			return nil, fmt.Errorf("grpc: register function is nil")
		}
		return NewGrpcHandler(conf, params.GrpcRegisterFunc, params.ExtraGrpcInterceptors...)
	case serverconf.FrameworkConnect:
		if params.ConnectRegisterMux == nil {
			return nil, fmt.Errorf("connect: register mux is nil")
		}
		return NewConnectHandler(conf, params.ConnectRegisterMux)
	default:
		return nil, fmt.Errorf("unsupported rpc type: %s", conf.FrameworkType)
	}
}
