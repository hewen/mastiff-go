// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/mock"
)

// MockGnetEventHandler is a mock implementation of gnet.EventHandler for testing.
type MockGnetEventHandler struct {
	mock.Mock
}

// OnBoot is called once when the engine starts.
func (m *MockGnetEventHandler) OnBoot(eng gnet.Engine) gnet.Action {
	args := m.Called(eng)
	return args.Get(0).(gnet.Action)
}

// OnShutdown is called when the engine is shutting down.
func (m *MockGnetEventHandler) OnShutdown(eng gnet.Engine) {
	m.Called(eng)
}

// OnOpen is called when a new connection is opened.
func (m *MockGnetEventHandler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	args := m.Called(c)
	return args.Get(0).([]byte), args.Get(1).(gnet.Action)
}

// OnClose is called when a connection is closed.
func (m *MockGnetEventHandler) OnClose(c gnet.Conn, err error) gnet.Action {
	args := m.Called(c, err)
	return args.Get(0).(gnet.Action)
}

// OnTick is called periodically by the event loop.
func (m *MockGnetEventHandler) OnTick() (time.Duration, gnet.Action) {
	args := m.Called()
	return args.Get(0).(time.Duration), args.Get(1).(gnet.Action)
}

// OnTraffic is triggered when a complete message is received.
func (m *MockGnetEventHandler) OnTraffic(c gnet.Conn) gnet.Action {
	args := m.Called(c)
	return args.Get(0).(gnet.Action)
}
