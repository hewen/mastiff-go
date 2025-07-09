package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGRPCAuthInterceptor(t *testing.T) {
	conf := Config{
		HeaderKey:     "authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList: []string{
			"/auth.Public/Login",
		},
	}

	interceptor := GRPCAuthInterceptor(conf)
	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	handlerCheckAuth := func(ctx context.Context, _ any) (any, error) {
		authInfo, ok := GetAuthInfoFromContext(ctx)
		if !ok || authInfo.UserID != "123" {
			return nil, status.Error(codes.Unauthenticated, "auth info missing or invalid")
		}

		return "ok", nil
	}

	t.Run("white list should pass", func(t *testing.T) {
		testGRPCWhiteList(t, interceptor, handler)
	})

	t.Run("missing metadata", func(t *testing.T) {
		testGRPCMissingMetadata(t, interceptor, handler)
	})

	t.Run("missing token", func(t *testing.T) {
		testGRPCMissingToken(t, interceptor, handler)
	})

	t.Run("invalid token", func(t *testing.T) {
		testGRPCInvalidToken(t, interceptor, handler)
	})

	t.Run("ok", func(t *testing.T) {
		testGRPCValidToken(t, interceptor, handlerCheckAuth, conf)
	})
}

func testGRPCWhiteList(t *testing.T, interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) {
	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Public/Login",
	}, handler)

	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func testGRPCMissingMetadata(t *testing.T, interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) {
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Secure/Action",
	}, handler)

	st, _ := status.FromError(err)
	assert.Equal(t, "missing metadata", st.Message())
}

func testGRPCMissingToken(t *testing.T, interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Secure/Action",
	}, handler)

	st, _ := status.FromError(err)
	assert.Equal(t, "missing token", st.Message())
}

func testGRPCInvalidToken(t *testing.T, interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) {
	md := metadata.Pairs("authorization", "Bearer invalid-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Secure/Action",
	}, handler)

	st, _ := status.FromError(err)
	assert.Equal(t, "invalid token", st.Message())
}

func testGRPCValidToken(t *testing.T, interceptor grpc.UnaryServerInterceptor, handler grpc.UnaryHandler, conf Config) {
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, conf.JWTSecret, time.Minute)
	md := metadata.Pairs("authorization", "Bearer "+tk)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Secure/Action",
	}, handler)

	assert.Nil(t, err)
	assert.Equal(t, "ok", resp)
}
