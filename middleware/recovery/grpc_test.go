package recovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMiddlewarePanic(t *testing.T) {
	handle := UnaryRecoveryInterceptor()
	_, err := handle(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		panic("test")
	})
	assert.Nil(t, err)
}
