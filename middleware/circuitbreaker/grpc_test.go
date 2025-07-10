// Package circuitbreaker provides a circuit breaker middleware for gRPC servers.
package circuitbreaker

import (
	"context"
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func TestUnaryCircuitBreakerInterceptor_Success(t *testing.T) {
	cfg := &Config{
		MaxRequests: 3,
		Interval:    1,
		Timeout:     1,
		ReadyToTrip: func(_ gobreaker.Counts) bool {
			return false
		},
	}
	mgr := NewManager(cfg)

	interceptor := UnaryCircuitBreakerInterceptor(mgr)

	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestUnaryCircuitBreakerInterceptor_Failure(t *testing.T) {
	cfg := &Config{
		MaxRequests: 1,
		Interval:    1,
		Timeout:     1,
		ReadyToTrip: func(_ gobreaker.Counts) bool {
			return true
		},
	}
	mgr := NewManager(cfg)
	mgr.Break("/fail", 1)

	interceptor := UnaryCircuitBreakerInterceptor(mgr)

	handler := func(_ context.Context, _ any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/fail"}, handler)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, status.Convert(err).Message(), "circuit breaker triggered")
}
