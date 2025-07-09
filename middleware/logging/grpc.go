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
		pr, _ := peer.FromContext(ctx)

		ctx = logger.NewOutgoingContextWithIncomingContext(ctx)
		l := logger.NewLoggerWithContext(ctx)

		resp, err := handler(ctx, req)

		logger.LogRequest(
			l,
			0,
			time.Since(start),
			getPeerIP(pr),
			info.FullMethod,
			"GRPC-GO-SERVER",
			req,
			resp,
			err,
		)
		return resp, err
	}
}

func getPeerIP(pr *peer.Peer) string {
	if pr != nil {
		return pr.Addr.String()
	}
	return ""
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
