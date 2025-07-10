// Package circuitbreaker provides a circuit breaker middleware for gRPC servers.
package circuitbreaker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// executeWithBreaker executes the function with circuit breaker.
func executeWithBreaker(
	mgr *Manager,
	method string,
	fn func() (any, error),
) (any, error) {
	breaker := mgr.Get(method)
	result, err := breaker.Execute(fn)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "circuit breaker triggered: %v", err)
	}
	return result, nil
}

// UnaryServerInterceptor returns a unary server interceptor with circuit breaker.
func UnaryServerInterceptor(mgr *Manager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		return executeWithBreaker(mgr, info.FullMethod, func() (any, error) {
			return handler(ctx, req)
		})
	}
}

// StreamServerInterceptor returns a stream server interceptor with circuit breaker.
func StreamServerInterceptor(mgr *Manager) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		_, err := executeWithBreaker(mgr, info.FullMethod, func() (any, error) {
			return nil, handler(srv, ss)
		})
		return err
	}
}
