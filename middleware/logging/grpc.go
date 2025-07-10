package logging

import (
	"context"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// UnaryServerInterceptor is a gRPC unary interceptor that logs the request and response details, including execution time and any errors.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
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
			"GRPC-GO-UNARY",
			req,
			resp,
			err,
		)
		return resp, err
	}
}

// StreamServerInterceptor is a gRPC stream interceptor that logs the request and response stream details, including execution time and any errors.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		ctx := logger.NewOutgoingContextWithIncomingContext(ss.Context())
		wrapped := &middleware.GrpcServerStream{
			ServerStream: ss,
			Ctx:          ctx,
		}

		err := handler(srv, wrapped)

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
			"GRPC-GO-STREAM",
			nil,
			nil,
			err,
		)

		return err
	}
}

// UnaryClientInterceptor creates a gRPC client interceptor that logs the request and response details, including execution time and any errors.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
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
