package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	handler := func(_ context.Context, _ any) (any, error) {
		time.Sleep(10 * time.Millisecond)
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/package.Service/TestMethod"}

	_, err := interceptor(context.Background(), nil, info, handler)
	assert.NoError(t, err)

	count := testutil.CollectAndCount(GRPCDuration)
	assert.Greater(t, count, 0, "Expected GRPCDuration to have collected a metric")
}

func TestStreamServerInterceptor(t *testing.T) {
	interceptor := StreamServerInterceptor()
	handler := func(_ any, _ grpc.ServerStream) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	info := &grpc.StreamServerInfo{FullMethod: "/package.Service/StreamMethod"}

	stream := &mockStream{}

	err := interceptor(nil, stream, info, handler)
	assert.NoError(t, err)

	count := testutil.CollectAndCount(GRPCDuration)
	assert.Greater(t, count, 0, "Expected GRPCDuration to have collected a metric")
}

// mockStream implements grpc.ServerStream for test.
type mockStream struct {
	grpc.ServerStream
}
