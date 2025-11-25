package auth

import (
	"context"

	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware/internal/shared"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// authenticate handles token extraction and validation.
func authenticate(ctx context.Context, method string, conf authconf.Config) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	token := extractTokenFromGrpcMetadata(md, conf.HeaderKey, conf.TokenPrefixes)
	if token != "" {
		authInfo, err := validateJWTToken(token, conf.JWTSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		logger.NewLoggerWithContext(ctx).Infof("auth info: %v", authInfo.Claims)
		ctx = contextkeys.SetAuthInfo(ctx, authInfo)
		ctx = contextkeys.SetUserID(ctx, authInfo.UserID)
		return ctx, nil
	}

	if isWhiteListed(method, conf.WhiteList) {
		return ctx, nil
	} else {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}
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
		srv any,
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
