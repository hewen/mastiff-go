package socketx

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/socketx/codec"
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

func (m *MockSocketHandler) UnbindDevice(_ string) {}

func (m *MockSocketHandler) PushTo(deviceID string, msg codec.Message, callback handler.AsyncCallback) error {
	args := m.Called(deviceID, msg, callback)
	return args.Error(0)
}

type testEventHandler struct{}

func (*testEventHandler) OnHandshakeMessage(c gnet.Conn, data codec.Message) codec.Message {
	_, _ = c.Write(data.GetPayload())
	return &codec.MQTTMessage{
		Payload: []byte("hello"),
		Header:  map[string]string{"clientID": "dev123", "topic": "test"},
	}
}
func (*testEventHandler) OnMessage(c gnet.Conn, data codec.Message) codec.Message {
	_, _ = c.Write(data.GetPayload())
	return &codec.MQTTMessage{
		Payload: []byte("hello"),
		Header:  map[string]string{"clientID": "dev123", "topic": "test"},
	}
}
func (*testEventHandler) MatchHandshakePrefix(_ []byte) (string, bool) {
	return "mqtt", true
}
func (*testEventHandler) NewCodec(_ string) (codec.SecureCodec, error) {
	return &codec.MQTTCodec{}, nil
}
func (*testEventHandler) OnBoot(_ gnet.Engine) (action gnet.Action) {
	return gnet.None
}
func (*testEventHandler) OnShutdown(_ gnet.Engine) {}
func (*testEventHandler) OnOpen(_ gnet.Conn) ([]byte, gnet.Action) {
	return nil, gnet.None
}
func (*testEventHandler) OnClose(_ gnet.Conn, _ error) gnet.Action {
	return gnet.None
}
func (*testEventHandler) OnTick() (time.Duration, gnet.Action) {
	return time.Second, gnet.None
}

func TestSocketServer_MQTT(t *testing.T) {
	port, _ := util.GetFreePort()
	addr := fmt.Sprintf("localhost:%d", port)
	protocol := "tcp"
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          protocol + "://" + addr,
	}

	s, err := NewSocketServer(conf, handler.BuildParams{
		GnetHandler: &testEventHandler{},
	})
	assert.NoError(t, err)

	go s.Start()
	defer s.Stop()

	err = waitForServer(addr, 2*time.Second)
	assert.NoError(t, err)

	assert.Equal(t, "gnet", s.Name())

	l := logger.NewLogger()
	s.WithLogger(l)

	conn, err := net.Dial(protocol, addr)
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	connectPacket := codec.BuildMQTTConnectPacket("dev123")
	_, err = conn.Write(connectPacket)
	assert.NoError(t, err)

	resp := make([]byte, 1)
	_, err = conn.Read(resp)
	assert.NoError(t, err)

	msg := &codec.MQTTMessage{
		Payload: []byte("hello"),
		Header:  map[string]string{"clientID": "dev123", "topic": "test"},
	}
	var wg sync.WaitGroup
	wg.Add(1)

	err = s.PushTo("dev123", msg, func(_ handler.Conn, writeErr error) error {
		defer wg.Done()
		assert.NoError(t, writeErr)
		return nil
	})
	assert.NoError(t, err)

	wg.Wait()

	s.UnbindDevice("dev123")

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
	server.Start()

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
