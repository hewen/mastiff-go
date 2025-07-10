// Package timeout provides a gRPC interceptor that sets a timeout for each request.
package timeout

import (
	"context"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptor(t *testing.T) {
	fn := UnaryServerInterceptor(time.Second)

	ctx := context.TODO()
	_, err := fn(
		ctx,
		nil,
		&grpc.UnaryServerInfo{
			FullMethod: "test",
		},
		func(_ context.Context, _ any) (any, error) {
			return nil, nil
		},
	)
	assert.Nil(t, err)
}

func TestStreamServerInterceptor(t *testing.T) {
	fn := StreamServerInterceptor(time.Second)

	ctx := context.TODO()
	err := fn(
		nil,
		&middleware.GrpcServerStream{Ctx: ctx},
		&grpc.StreamServerInfo{
			FullMethod: "test",
		},
		func(_ any, _ grpc.ServerStream) error {
			return nil
		},
	)
	assert.Nil(t, err)
}
