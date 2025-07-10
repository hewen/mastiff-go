// Package recovery provides a gRPC interceptor that recovers from panics.
package recovery

import (
	"context"
	"testing"

	"github.com/hewen/mastiff-go/middleware/internal/shared"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptor(t *testing.T) {
	fn := UnaryServerInterceptor()
	_, err := fn(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		panic("test")
	})
	assert.Nil(t, err)
}

func TestStreamServerInterceptor(t *testing.T) {
	fn := StreamServerInterceptor()

	ctx := context.TODO()
	err := fn(
		nil,
		&shared.GrpcServerStream{Ctx: ctx},
		&grpc.StreamServerInfo{
			FullMethod: "test",
		},
		func(_ any, _ grpc.ServerStream) error {
			return nil
		},
	)
	assert.Nil(t, err)
}
