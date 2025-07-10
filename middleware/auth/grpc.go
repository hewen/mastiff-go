package auth

import (
	"context"

	"github.com/hewen/mastiff-go/internal/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor is a grpc interceptor for authentication and authorization.
func UnaryServerInterceptor(conf Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isWhiteListed(info.FullMethod, conf.WhiteList) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := extractTokenFromGrpcMetadata(md, conf.HeaderKey, conf.TokenPrefixes)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		authInfo, err := validateJWTToken(token, conf.JWTSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = contextkeys.SetAuthInfo(ctx, authInfo)
		return handler(ctx, req)
	}
}
