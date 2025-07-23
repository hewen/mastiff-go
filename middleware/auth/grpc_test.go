package auth

import (
	"context"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/middleware/internal/shared"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var testConf = authconf.Config{
	JWTSecret:     "test-secret",
	HeaderKey:     "authorization",
	TokenPrefixes: []string{"Bearer"},
	WhiteList:     []string{"/TestService/Public"},
}

func TestUnaryServerInterceptor_WhiteList(t *testing.T) {
	interceptor := UnaryServerInterceptor(testConf)
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/TestService/Public"}

	resp, err := interceptor(
		ctx,
		"req",
		info,
		func(_ context.Context, _ interface{}) (interface{}, error) {
			return "ok", nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestUnaryServerInterceptor_MissingMetadata(t *testing.T) {
	interceptor := UnaryServerInterceptor(testConf)
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/TestService/Private"}

	_, err := interceptor(ctx, "req", info, nil)

	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnaryServerInterceptor_MissingToken(t *testing.T) {
	interceptor := UnaryServerInterceptor(testConf)
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/TestService/Private"}

	_, err := interceptor(ctx, "req", info, nil)

	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnaryServerInterceptor_InvalidToken(t *testing.T) {
	interceptor := UnaryServerInterceptor(testConf)
	md := metadata.New(map[string]string{"authorization": "Bearer bad-token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/TestService/Private"}

	_, err := interceptor(ctx, "req", info, nil)

	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnaryServerInterceptor_ValidToken(t *testing.T) {
	interceptor := UnaryServerInterceptor(testConf)
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, testConf.JWTSecret, time.Minute)
	md := metadata.Pairs("authorization", "Bearer "+tk)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/TestService/Private"}

	called := false
	resp, err := interceptor(ctx, "req", info, func(ctx context.Context, _ interface{}) (interface{}, error) {
		authInfo, _ := contextkeys.GetAuthInfo(ctx)
		assert.NotNil(t, authInfo)
		assert.Equal(t, "123", authInfo.UserID)
		called = true
		return "success", nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "success", resp)
}

func TestStreamServerInterceptor_ValidToken(t *testing.T) {
	interceptor := StreamServerInterceptor(testConf)
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, testConf.JWTSecret, time.Minute)
	md := metadata.Pairs("authorization", "Bearer "+tk)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	stream := &shared.GrpcServerStream{Ctx: ctx}
	info := &grpc.StreamServerInfo{FullMethod: "/TestService/Private"}

	called := false
	err := interceptor(nil, stream, info, func(_ any, ss grpc.ServerStream) error {
		authInfo, _ := contextkeys.GetAuthInfo(ss.Context())
		assert.NotNil(t, authInfo)
		assert.Equal(t, "123", authInfo.UserID)
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}
