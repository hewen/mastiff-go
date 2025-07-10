// Package ratelimit provides a rate limiter middleware.
package ratelimit

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryRateLimitInterceptor creates a gRPC unary interceptor with rate limiter.
func UnaryRateLimitInterceptor(mgr *LimiterManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		route := info.FullMethod
		cfg := mgr.config.PerRoute[route]
		if cfg == nil {
			cfg = mgr.config.Default
		}
		if cfg == nil {
			return handler(ctx, req)
		}
		key := mgr.getKeyFromContext(ctx, route, cfg)
		limiter := mgr.getOrCreateLimiter(key, cfg)
		if err := limiter.AllowOrWait(ctx); err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}
