package logging

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
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

func TestGetPeerIP(t *testing.T) {
	t.Run("peer is nil", func(t *testing.T) {
		ip := getPeerIP(nil)
		assert.Equal(t, "", ip)
	})

	t.Run("peer is not nil", func(t *testing.T) {
		addr := &net.TCPAddr{
			IP:   net.ParseIP("192.168.1.100"),
			Port: 8080,
		}
		pr := &peer.Peer{
			Addr: addr,
		}
		ip := getPeerIP(pr)
		assert.Equal(t, "192.168.1.100:8080", ip) // 因为 Addr.String() 返回 ip:port
	})
}
