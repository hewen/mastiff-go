package handler

import (
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/server/socketx/codec"
	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGnetEventHandler struct {
	mock.Mock
}

func (m *MockGnetEventHandler) OnBoot(eng gnet.Engine) gnet.Action {
	args := m.Called(eng)
	return args.Get(0).(gnet.Action)
}

func (m *MockGnetEventHandler) OnShutdown(eng gnet.Engine) {
	m.Called(eng)
}

func (m *MockGnetEventHandler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	args := m.Called(c)
	return args.Get(0).([]byte), args.Get(1).(gnet.Action)
}

func (m *MockGnetEventHandler) OnClose(c gnet.Conn, err error) gnet.Action {
	args := m.Called(c, err)
	return args.Get(0).(gnet.Action)
}

func (m *MockGnetEventHandler) OnTick() (time.Duration, gnet.Action) {
	args := m.Called()
	return args.Get(0).(time.Duration), args.Get(1).(gnet.Action)
}

func (m *MockGnetEventHandler) OnHandshakeMessage(c gnet.Conn, data codec.Message) codec.Message {
	args := m.Called(c, data)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(codec.Message)
}

func (m *MockGnetEventHandler) OnMessage(c gnet.Conn, data codec.Message) codec.Message {
	args := m.Called(c, data)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(codec.Message)
}

func (m *MockGnetEventHandler) MatchHandshakePrefix(data []byte) (string, bool) {
	args := m.Called(data)
	return args.String(0), args.Bool(1)
}

func (m *MockGnetEventHandler) NewCodec(protocol string) (codec.SecureCodec, error) {
	args := m.Called(protocol)
	return args.Get(0).(codec.SecureCodec), args.Error(1)
}

// nolint
func TestNewHandler(t *testing.T) {
	tests := []struct {
		conf        *serverconf.SocketConfig
		params      BuildParams
		name        string
		errorMsg    string
		expectError bool
	}{
		{
			name:        "nil config",
			conf:        nil,
			params:      BuildParams{},
			expectError: true,
			errorMsg:    "empty socket conf",
		},
		{
			name: "gnet framework with nil handler",
			conf: &serverconf.SocketConfig{
				FrameworkType: serverconf.FrameworkGnet,
				Addr:          ":8080",
			},
			params: BuildParams{
				GnetHandler: nil,
			},
			expectError: true,
			errorMsg:    "gnet: handler is nil",
		},
		{
			name: "gnet framework with valid handler",
			conf: &serverconf.SocketConfig{
				FrameworkType: serverconf.FrameworkGnet,
				Addr:          ":8080",
				MaxIdleTime:   30 * time.Second,
				TickInterval:  10 * time.Second,
			},
			params: BuildParams{
				GnetHandler: new(MockGnetEventHandler),
			},
			expectError: false,
		},
		{
			name: "unsupported framework type",
			conf: &serverconf.SocketConfig{
				FrameworkType: "unsupported",
				Addr:          ":8080",
			},
			params: BuildParams{
				GnetHandler: new(MockGnetEventHandler),
			},
			expectError: true,
			errorMsg:    "unsupported socket type: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler(tt.conf, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.Equal(t, "gnet", handler.Name())
			}
		})
	}
}

func TestBuildParams(t *testing.T) {
	mockHandler := new(MockGnetEventHandler)

	params := BuildParams{
		GnetHandler: mockHandler,
	}

	assert.Equal(t, mockHandler, params.GnetHandler)
}

func TestErrorConstants(t *testing.T) {
	assert.NotNil(t, ErrEmptySocketConf)
	assert.Equal(t, "empty socket conf", ErrEmptySocketConf.Error())
}

func TestNewHandler_GnetIntegration(t *testing.T) {
	// Test that the returned handler implements SocketHandler interface
	mockEventHandler := new(MockGnetEventHandler)

	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
		MaxIdleTime:   30 * time.Second,
		TickInterval:  10 * time.Second,
	}

	params := BuildParams{
		GnetHandler: mockEventHandler,
	}

	handler, err := NewHandler(conf, params)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Test interface methods
	assert.Equal(t, "gnet", handler.Name())

	// Test PushTo with non-existent device
	mockMessage := &MockMessage{}
	mockMessage.On("GetPayload").Return([]byte("test"))
	mockMessage.On("GetHeader").Return(map[string]string{})

	err = handler.PushTo("nonexistent", mockMessage, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrDeviceNotFound, err)
}

type MockMessage struct {
	mock.Mock
}

func (m *MockMessage) GetPayload() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockMessage) GetHeader() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func TestNewHandler_ConfigValidation(t *testing.T) {
	mockEventHandler := new(MockGnetEventHandler)

	// Test with minimal config
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":0", // Use port 0 for testing
	}

	params := BuildParams{
		GnetHandler: mockEventHandler,
	}

	handler, err := NewHandler(conf, params)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Verify the handler was created with the config
	gnetHandler, ok := handler.(*GnetHandler)
	assert.True(t, ok)
	assert.Equal(t, conf, gnetHandler.conf)
	assert.Equal(t, mockEventHandler, gnetHandler.event)
	assert.NotNil(t, gnetHandler.connManager)
	assert.NotNil(t, gnetHandler.logger)
}

func TestNewHandler_AllFrameworkTypes(t *testing.T) {
	mockEventHandler := new(MockGnetEventHandler)

	// Test all known framework types
	frameworkTests := []struct {
		framework   serverconf.FrameworkType
		expectError bool
	}{
		{serverconf.FrameworkGnet, false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range frameworkTests {
		t.Run(string(tt.framework), func(t *testing.T) {
			conf := &serverconf.SocketConfig{
				FrameworkType: tt.framework,
				Addr:          ":8080",
			}

			params := BuildParams{
				GnetHandler: mockEventHandler,
			}

			handler, err := NewHandler(conf, params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}
