package logging

import (
	"context"
	"testing"

	"github.com/hewen/mastiff-go/middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptor(t *testing.T) {
	fn := UnaryServerInterceptor()
	resp, err := fn(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return "test", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", resp)
}

func TestStreamServerInterceptor(t *testing.T) {
	fn := StreamServerInterceptor()
	err := fn(
		nil,
		&middleware.GrpcServerStream{Ctx: context.Background()},
		&grpc.StreamServerInfo{
			FullMethod: "test",
		},
		func(_ any, _ grpc.ServerStream) error {
			return nil
		},
	)
	assert.Nil(t, err)
}

func TestUnaryClientInterceptor(t *testing.T) {
	fn := UnaryClientInterceptor()
	err := fn(
		context.TODO(),
		"test",
		"req",
		"reply",
		&grpc.ClientConn{},
		func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return nil
		},
	)
	assert.Nil(t, err)
}
