package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGrpcServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)
	enableMetrics := true
	c := &serverconf.RPCConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Middlewares: middlewareconf.Config{
			EnableMetrics: &enableMetrics,
		},
		Reflection: true,
	}

	s, err := NewGrpcHandler(
		c,
		func(_ *grpc.Server) {
			// not doing
		},
		testInterceptor(),
	)

	assert.NotNil(t, s)
	assert.Nil(t, err)

	_ = s.Name()

	go func() {
		defer func() { _ = s.Stop() }()
		_ = s.Start()
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

func TestGrpcHandler_StartError(t *testing.T) {
	grpcServer := &GrpcHandler{
		s:    grpc.NewServer(),
		ln:   &brokenListener{},
		addr: "mock",
	}

	err := grpcServer.Start()
	assert.NotNil(t, err)
}

func testInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
}

func TestGrpcServerStop(t *testing.T) {
	c := &serverconf.RPCConfig{}

	s, err := NewGrpcHandler(
		c,
		func(_ *grpc.Server) {
			// not doing
		},
		testInterceptor(),
	)

	assert.NotNil(t, s)
	assert.Nil(t, err)
	err = s.Stop()
	assert.Nil(t, err)
}

func TestGrpcServerEmptyConfig(t *testing.T) {
	_, err := NewGrpcHandler(
		nil,
		func(_ *grpc.Server) {
			// not doing
		},
		testInterceptor(),
	)

	assert.EqualValues(t, err, ErrEmptyRPCConf)
}

func TestNewGrpcServerError(t *testing.T) {
	c := &serverconf.RPCConfig{
		Addr: "error",
	}

	_, err := NewGrpcHandler(
		c,
		func(_ *grpc.Server) {
			// not doing
		},
		testInterceptor(),
	)

	assert.EqualValues(t, "listen tcp: address error: missing port in address", err.Error())
}
