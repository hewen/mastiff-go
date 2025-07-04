package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGrpcServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)
	c := &GrpcConf{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	s, err := NewGrpcServer(c, func(_ *grpc.Server) {
		// not doing
	})
	assert.NotNil(t, s)
	assert.Nil(t, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

func testInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
}

func TestGrpcServerStop(t *testing.T) {
	c := &GrpcConf{}

	s, err := NewGrpcServer(c, func(_ *grpc.Server) {
		// not doing
	}, testInterceptor())
	assert.NotNil(t, s)
	assert.Nil(t, err)
	s.Stop()
}

func TestMiddleware(t *testing.T) {
	gs := GrpcServer{}
	_, err := gs.middleware(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})
	assert.Nil(t, err)
}

func TestMiddlewarePanic(t *testing.T) {
	gs := GrpcServer{}
	_, err := gs.middleware(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		panic("test")
	})
	assert.Equal(t, ErrGrpcExecPanic, err)
}

func TestTimeoutInterceptor(t *testing.T) {
	gs := GrpcServer{}
	fn := gs.timeoutInterceptor(0)

	ctx := context.TODO()
	_, err := fn(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})
	assert.Nil(t, err)
}

func TestGrpcServerEmptyConfig(t *testing.T) {
	_, err := NewGrpcServer(nil, func(_ *grpc.Server) {
		// not doing
	})
	assert.EqualValues(t, err, ErrEmptyGrpcConf)
}
