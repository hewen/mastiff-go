package socketx

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/socketx/handler"
	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSocketHandler struct {
	mock.Mock
}

func (m *MockSocketHandler) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSocketHandler) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSocketHandler) Name() string {
	args := m.Called()
	return args.String(0)
}

func TestSocketServer(t *testing.T) {
	port, _ := util.GetFreePort()
	addr := fmt.Sprintf("localhost:%d", port)
	protocol := "tcp"
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          protocol + "://" + addr,
	}

	mockEvent := new(handler.MockGnetEventHandler)
	mockEvent.On("OnBoot", mock.Anything).Return(gnet.None)
	mockEvent.On("OnShutdown", mock.Anything).Return()
	expectedData := []byte("response data")
	mockEvent.On("OnOpen", mock.Anything).Return(expectedData, gnet.None)
	mockEvent.On("OnClose", mock.Anything, mock.Anything).Return(gnet.None)
	mockEvent.On("OnTraffic", mock.Anything).Return(gnet.None)

	s, err := NewSocketServer(conf, handler.BuildParams{
		GnetHandler: mockEvent,
	})
	assert.NoError(t, err)

	go s.Start()

	err = waitForServer(addr, 2*time.Second)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("socket gnet server(%s)", conf.Addr), s.Name())

	l := logger.NewLogger()
	s.WithLogger(l)

	conn, err := net.Dial(protocol, addr)
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	req := []byte("test")
	_, err = conn.Write(req)
	assert.Nil(t, err)

	resp := make([]byte, len(req))
	_, err = conn.Read(resp)
	assert.Nil(t, err)
}

func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			return conn.Close()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("server not ready in %s", timeout)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSocketServer_NewSocketServer_Error(t *testing.T) {
	s, err := NewSocketServer(&serverconf.SocketConfig{}, handler.BuildParams{})
	assert.Nil(t, s)
	assert.Error(t, err)
}

func TestSocketServer_Start_Error(t *testing.T) {
	// Create a mock handler that returns an error on Start
	mockHandler := &MockSocketHandler{}

	server := &SocketServer{
		handler: mockHandler,
		logger:  logger.NewLogger(),
	}

	// Set up expectation for Start to return an error
	expectedError := errors.New("start failed")
	mockHandler.On("Start").Return(expectedError)

	// Test Start - should not panic even when handler.Start() returns error
	assert.Panics(t, func() {
		server.Start()
	})

	// Verify that Start was called
	mockHandler.AssertExpectations(t)
}

func TestSocketServer_Stop(t *testing.T) {
	// Create a mock handler
	mockHandler := &MockSocketHandler{}

	server := &SocketServer{
		handler: mockHandler,
		logger:  logger.NewLogger(),
	}

	// Set up expectation for Stop
	mockHandler.On("Stop").Return(nil)

	// Test Stop
	server.Stop()

	// Verify that Stop was called
	mockHandler.AssertExpectations(t)
}

func TestSocketServer_Stop_WithError(t *testing.T) {
	// Create a mock handler that returns an error on Stop
	mockHandler := &MockSocketHandler{}

	server := &SocketServer{
		handler: mockHandler,
		logger:  logger.NewLogger(),
	}

	// Set up expectation for Stop to return an error
	expectedError := errors.New("stop failed")
	mockHandler.On("Stop").Return(expectedError)

	// Test Stop - should not panic even when handler.Stop() returns error
	server.Stop()

	// Verify that Stop was called
	mockHandler.AssertExpectations(t)
}
