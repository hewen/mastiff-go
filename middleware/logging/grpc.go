package logging

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// UnaryLoggingInterceptor is a gRPC interceptor that logs the request and response details, including execution time and any errors.
func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		ctx = logger.NewOutgoingContextWithIncomingContext(ctx)

		resp, err := handler(ctx, req)

		l := logger.NewLoggerWithContext(ctx)
		var ip string
		if pr, _ := peer.FromContext(ctx); pr != nil {
			ip = pr.Addr.String()
		}
		logger.LogRequest(
			l,
			0,
			time.Since(start),
			ip,
			info.FullMethod,
			"GRPC-GO-SERVER",
			req,
			resp,
			err,
		)
		return resp, err
	}
}

// UnaryClientLoggingInterceptor creates a gRPC client interceptor that logs the request and response details, including execution time and any errors.
func UnaryClientLoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		l := logger.NewLoggerWithContext(ctx)
		err := invoker(ctx, method, req, reply, cc, opts...)

		logger.LogRequest(
			l,
			0,
			time.Since(start),
			cc.Target(),
			method,
			"GRPC-GO-CLIENT",
			req,
			reply,
			err,
		)

		return err
	}
}
