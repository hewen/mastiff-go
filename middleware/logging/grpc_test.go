package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryLoggingInterceptor(t *testing.T) {
	handle := UnaryLoggingInterceptor()
	resp, err := handle(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return "test", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", resp)
}

func TestUnaryClientLoggingInterceptor(t *testing.T) {
	handle := UnaryClientLoggingInterceptor()
	err := handle(
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
