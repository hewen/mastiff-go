package auth

import (
	"context"

	"github.com/hewen/mastiff-go/config/middleware/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/middleware/internal/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// authenticate handles token extraction and validation.
func authenticate(ctx context.Context, method string, conf authconf.Config) (context.Context, error) {
	if isWhiteListed(method, conf.WhiteList) {
		return ctx, nil
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

	newCtx := contextkeys.SetAuthInfo(ctx, authInfo)
	return newCtx, nil
}

// UnaryServerInterceptor implements unary auth interceptor.
func UnaryServerInterceptor(conf authconf.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		newCtx, err := authenticate(ctx, info.FullMethod, conf)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor implements stream auth interceptor.
func StreamServerInterceptor(conf authconf.Config) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		newCtx, err := authenticate(ss.Context(), info.FullMethod, conf)
		if err != nil {
			return err
		}

		wrapped := &shared.GrpcServerStream{
			ServerStream: ss,
			Ctx:          newCtx,
		}
		return handler(srv, wrapped)
	}
}
