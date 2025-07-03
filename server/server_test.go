package server

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockServers struct{}

func (MockServers) Start() {}

func (MockServers) Stop() {}

func TestServersStart(_ *testing.T) {
	var servers Servers
	ms := MockServers{}
	servers.Add(ms)
	servers.Start()
}

func TestServersStop(_ *testing.T) {
	var servers Servers
	ms := MockServers{}
	servers.Add(ms)
	servers.Stop()
}

func TestGracefulStop(t *testing.T) {
	var called bool
	var mu sync.Mutex

	stopFunc := []func(){
		func() {
			mu.Lock()
			called = true
			mu.Unlock()
		},
	}

	for i := range stopFunc {
		AddGracefulStop(stopFunc[i])
	}

	gracefulStop()

	p, err := os.FindProcess(os.Getpid())
	assert.Nil(t, err)

	err = p.Signal(os.Interrupt)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, true, called)
	mu.Unlock()
}
