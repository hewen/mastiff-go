// Package timeout provides a gRPC interceptor that sets a timeout for each request.
package timeout

import (
	"context"
	"time"

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
