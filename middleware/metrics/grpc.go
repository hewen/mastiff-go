// Package metrics provides middleware for recording request metrics.
package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor is a gRPC unary interceptor for recording request metrics.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		st, _ := status.FromError(err)

		service, method := splitMethod(info.FullMethod)
		GRPCDuration.WithLabelValues(service, method, st.Code().String()).Observe(duration)

		return resp, err
	}
}

// StreamServerInterceptor is a gRPC stream interceptor for recording request metrics.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start).Seconds()

		st, _ := status.FromError(err)
		service, method := splitMethod(info.FullMethod)
		GRPCDuration.WithLabelValues(service, method, st.Code().String()).Observe(duration)

		return err
	}
}

// splitMethod splits a full method into service and method.
func splitMethod(fullMethod string) (service, method string) {
	// fullMethod: "/package.Service/Method"
	parts := []rune(fullMethod)
	lastSlash := 0
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			lastSlash = i
			break
		}
	}
	return string(parts[1:lastSlash]), string(parts[lastSlash+1:])
}
