// Package circuitbreaker provides a circuit breaker middleware for gRPC servers.
package circuitbreaker

import (
	"context"
	"testing"

	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_Success(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 3,
		Interval:    1,
		Timeout:     1,
	}
	mgr := NewManager(cfg)

	interceptor := UnaryServerInterceptor(mgr)

	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestUnaryServerInterceptor_Failure(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                "consecutive_failures",
			ConsecutiveFailures: 1,
		},
	}
	mgr := NewManager(cfg)
	mgr.Break("/fail", 1)

	interceptor := UnaryServerInterceptor(mgr)

	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/fail"}, handler)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, status.Convert(err).Message(), "circuit breaker triggered")
}

func TestStreamServerInterceptor_Success(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 3,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                "consecutive_failures",
			ConsecutiveFailures: 1,
		},
	}
	mgr := NewManager(cfg)

	interceptor := StreamServerInterceptor(mgr)

	err := interceptor(context.Background(), nil, &grpc.StreamServerInfo{FullMethod: "/test"}, func(_ any, _ grpc.ServerStream) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestStreamServerInterceptor_Failure(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		Policy: &circuitbreakerconf.PolicyConfig{
			Type:                "consecutive_failures",
			ConsecutiveFailures: 1,
		},
	}
	mgr := NewManager(cfg)
	mgr.Break("/fail", 1)

	interceptor := StreamServerInterceptor(mgr)

	err := interceptor(context.Background(), nil, &grpc.StreamServerInfo{FullMethod: "/fail"}, func(_ any, _ grpc.ServerStream) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, status.Convert(err).Message(), "circuit breaker triggered")
}
