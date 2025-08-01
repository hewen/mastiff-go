package server

import (
	"sync"
	"testing"

	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
)

type MockServers struct{}

func (*MockServers) Start() {}

func (*MockServers) Stop() {}

func (*MockServers) Name() string {
	return "MOCK SERVER"
}

func (s *MockServers) WithLogger(_ logger.Logger) {
}

func TestServersStart(_ *testing.T) {
	var servers Servers
	ms := &MockServers{}
	servers.Add(ms)
	servers.Start()
	servers.Stop()
}

func TestServersStop(_ *testing.T) {
	var servers Servers
	ms := &MockServers{}
	servers.Add(ms)
	servers.Stop()
}

func TestGracefulStop(t *testing.T) {
	var mu sync.Mutex
	var called bool

	AddGracefulStop(func() {
		t.Log("stopFunc called")
		mu.Lock()
		called = true
		mu.Unlock()
	})

	shutdown()

	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}
