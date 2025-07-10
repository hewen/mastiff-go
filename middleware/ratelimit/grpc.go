package ratelimit

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// checkRateLimit checks the rate limit for the given route.
func checkRateLimit(ctx context.Context, mgr *LimiterManager, route string) error {
	cfg := mgr.config.PerRoute[route]
	if cfg == nil {
		cfg = mgr.config.Default
	}
	if cfg == nil {
		return nil
	}

	key := mgr.getKeyFromContext(ctx, route, cfg)
	limiter := mgr.getOrCreateLimiter(key, cfg)

	if err := limiter.AllowOrWait(ctx); err != nil {
		return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}
	return nil
}

// UnaryServerInterceptor creates a gRPC unary interceptor with rate limiter.
func UnaryServerInterceptor(mgr *LimiterManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := checkRateLimit(ctx, mgr, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor creates a gRPC stream interceptor with rate limiter.
func StreamServerInterceptor(mgr *LimiterManager) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := checkRateLimit(ss.Context(), mgr, info.FullMethod); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
