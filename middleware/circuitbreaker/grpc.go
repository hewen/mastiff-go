// Package circuitbreaker provides a circuit breaker middleware for gRPC servers.
package circuitbreaker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor creates a gRPC unary interceptor with circuit breaker.
func UnaryServerInterceptor(mgr *Manager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		breaker := mgr.Get(info.FullMethod)

		result, err := breaker.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})

		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "circuit breaker triggered: %v", err)
		}
		return result, nil
	}
}
