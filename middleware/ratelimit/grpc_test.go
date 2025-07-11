// Package ratelimit provides a rate limiter middleware for gRPC servers.
package ratelimit

import (
	"context"
	"testing"

	"github.com/hewen/mastiff-go/middleware/internal/shared"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUnaryServerInterceptor(t *testing.T) {
	cfg := &Config{
		PerRoute: map[string]*RouteLimitConfig{
			"/ratelimit.TestService/Ping": &RouteLimitConfig{
				Rate:        1,
				Burst:       1,
				Mode:        ModeAllow,
				EnableRoute: true,
				EnableIP:    true,
			},
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	handle := UnaryServerInterceptor(mgr)
	_, err := handle(context.Background(),
		&emptypb.Empty{},
		&grpc.UnaryServerInfo{
			FullMethod: "/ratelimit.TestService/Test",
		},
		func(_ context.Context, _ any) (any, error) {
			return &emptypb.Empty{}, nil
		},
	)
	assert.Nil(t, err)

	_, err = handle(context.Background(),
		&emptypb.Empty{},
		&grpc.UnaryServerInfo{
			FullMethod: "/ratelimit.TestService/Ping",
		},
		func(_ context.Context, _ any) (any, error) {
			return &emptypb.Empty{}, nil
		},
	)
	assert.Nil(t, err)

	_, err = handle(context.Background(),
		&emptypb.Empty{},
		&grpc.UnaryServerInfo{
			FullMethod: "/ratelimit.TestService/Ping",
		},
		func(_ context.Context, _ any) (any, error) {
			return &emptypb.Empty{}, nil
		},
	)
	assert.Equal(t, status.Errorf(codes.ResourceExhausted, "rate limit exceeded"), err)
}

func TestStreamServerInterceptor(t *testing.T) {
	cfg := &Config{
		Default: &RouteLimitConfig{
			Rate:        1,
			Burst:       1,
			Mode:        ModeAllow,
			EnableRoute: true,
			EnableIP:    true,
		},
	}
	mgr := NewLimiterManager(cfg)
	defer mgr.Stop()

	handle := StreamServerInterceptor(mgr)
	err := handle(
		nil,
		&shared.GrpcServerStream{Ctx: context.Background()},
		&grpc.StreamServerInfo{
			FullMethod: "/ratelimit.TestService/Ping",
		},
		func(_ any, _ grpc.ServerStream) error {
			return nil
		},
	)
	assert.Nil(t, err)

	err = handle(
		nil,
		&shared.GrpcServerStream{Ctx: context.Background()},
		&grpc.StreamServerInfo{
			FullMethod: "/ratelimit.TestService/Ping",
		},
		func(_ any, _ grpc.ServerStream) error {
			return nil
		},
	)
	assert.Equal(t, status.Errorf(codes.ResourceExhausted, "rate limit exceeded"), err)
}
