// Package timeout provides a gRPC interceptor that sets a timeout for each request.
package timeout

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/middleware"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor creates a gRPC interceptor that sets a timeout for each request.
func UnaryServerInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

// StreamServerInterceptor creates a gRPC stream interceptor that sets a timeout for each stream.
func StreamServerInterceptor(timeout time.Duration) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx, cancel := context.WithTimeout(ss.Context(), timeout)
		defer cancel()

		wrapped := &middleware.GrpcServerStream{
			ServerStream: ss,
			Ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}
