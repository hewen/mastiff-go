package server

import (
	"testing"
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
