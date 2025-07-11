package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/hewen/mastiff-go/middleware/recovery"
	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGrpcServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)
	enableMetrics := true
	c := &GrpcConf{
		Addr: fmt.Sprintf("localhost:%d", port),
		Middlewares: middleware.Config{
			EnableMetrics: &enableMetrics,
		},
		Reflection: true,
	}

	s, err := NewGrpcServer(c, func(_ *grpc.Server) {
		// not doing
	}, recovery.UnaryServerInterceptor())
	assert.NotNil(t, s)
	assert.Nil(t, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

type brokenListener struct{}

func (b *brokenListener) Accept() (net.Conn, error) {
	return nil, errors.New("mock accept error")
}
func (b *brokenListener) Close() error {
	return nil
}
func (b *brokenListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

func TestGrpcServer_StartError(_ *testing.T) {
	grpcServer := &GrpcServer{
		s:    grpc.NewServer(),
		l:    logger.NewLogger(),
		ln:   &brokenListener{},
		addr: "mock",
	}

	grpcServer.Start()
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

func TestGrpcServerEmptyConfig(t *testing.T) {
	_, err := NewGrpcServer(nil, func(_ *grpc.Server) {
		// not doing
	})
	assert.EqualValues(t, err, ErrEmptyGrpcConf)
}

func TestNewGrpcServerError(t *testing.T) {
	c := &GrpcConf{
		Addr: "error",
	}
	_, err := NewGrpcServer(c, func(_ *grpc.Server) {
		// not doing
	})
	assert.EqualValues(t, "listen tcp: address error: missing port in address", err.Error())
}
