// Package timeout provides a gRPC interceptor that sets a timeout for each request.
package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryTimeoutInterceptor creates a gRPC interceptor that sets a timeout for each request.
func UnaryTimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}
