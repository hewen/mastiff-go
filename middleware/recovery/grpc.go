// Package recovery provides a gRPC interceptor that recovers from panics.
package recovery

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/hewen/mastiff-go/logger"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor recovers from panics in unary handlers and logs the error.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer recoverAndLog(ctx)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor recovers from panics in stream handlers and logs the error.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		defer recoverAndLog(ss.Context())
		return handler(srv, ss)
	}
}

// recoverAndLog handles panic recovery and logs the stack trace.
func recoverAndLog(ctx context.Context) {
	if r := recover(); r != nil {
		l := logger.NewLoggerWithContext(ctx)
		l.Errorf("panic: %v $%s", r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
	}
}
