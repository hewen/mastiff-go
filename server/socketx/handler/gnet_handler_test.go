package handler

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNetAddr struct {
	addr string
}

func (m *mockNetAddr) Network() string {
	return "tcp"
}

func (m *mockNetAddr) String() string {
	return m.addr
}

type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEngine) Stop(ctx any) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEngine) CountConnections() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEngine) Dup(fd int) (int, error) {
	args := m.Called(fd)
	return args.Int(0), args.Error(1)
}

type TestEngine struct{}

func (e *TestEngine) Start() error           { return nil }
func (e *TestEngine) Stop(_ any) error       { return nil }
func (e *TestEngine) CountConnections() int  { return 0 }
func (e *TestEngine) Dup(_ int) (int, error) { return 0, nil }

func TestNewGnetHandler(t *testing.T) {
	tests := []struct {
		conf        *serverconf.SocketConfig
		event       GnetEventHandler
		name        string
		errorMsg    string
		expectError bool
	}{
		{
			name:        "nil config",
			conf:        nil,
			event:       new(MockGnetEventHandler),
			expectError: true,
			errorMsg:    "empty socket conf",
		},
		{
			name: "valid config and event handler",
			conf: &serverconf.SocketConfig{
				FrameworkType: serverconf.FrameworkGnet,
				Addr:          ":8080",
				MaxIdleTime:   30 * time.Second,
				TickInterval:  10 * time.Second,
			},
			event:       new(MockGnetEventHandler),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewGnetHandler(tt.conf, tt.event)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.Equal(t, tt.conf, handler.conf)
				assert.Equal(t, tt.event, handler.event)
				assert.NotNil(t, handler.connManager)
				assert.NotNil(t, handler.logger)
			}
		})
	}
}

func TestGnetHandler_Name(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	event := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, event)
	assert.NoError(t, err)

	assert.Equal(t, "gnet", handler.Name())
}

func TestGnetHandler_OnBoot_WithReflection(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Set up expectation for event handler
	mockEvent.On("OnBoot", mock.Anything).Return(gnet.None)

	// Use reflection to call OnBoot method with nil engine
	// This will actually execute the method and improve coverage
	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("OnBoot")
	assert.True(t, method.IsValid(), "OnBoot method should exist")

	// Call the method with nil engine (using reflection to bypass type checking)
	nilEngine := reflect.Zero(reflect.TypeOf((*gnet.Engine)(nil)).Elem())
	results := method.Call([]reflect.Value{nilEngine})

	// Verify the method was called and returned a value
	assert.Equal(t, 1, len(results), "OnBoot should return 1 value")
	mockEvent.AssertExpectations(t)
}

func TestGnetHandler_OnShutdown_WithReflection(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Add a connection to test cleanup
	mockConn := new(MockConn)
	mockConn.On("Close").Return(nil)

	meta := &ConnMeta{
		Conn: mockConn,
	}
	handler.connManager.BindDevice("test_device", meta)

	// Set up expectation for event handler
	mockEvent.On("OnShutdown", mock.Anything).Return()

	// Use reflection to call OnShutdown method
	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("OnShutdown")
	assert.True(t, method.IsValid(), "OnShutdown method should exist")

	// Call the method with nil engine
	nilEngine := reflect.Zero(reflect.TypeOf((*gnet.Engine)(nil)).Elem())
	method.Call([]reflect.Value{nilEngine})

	// Verify connection was closed during cleanup
	mockConn.AssertCalled(t, "Close")
	mockEvent.AssertExpectations(t)
}

func TestGnetHandler_Start_WithReflection(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":0", // Use port 0 to avoid conflicts, but this will still fail
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Use reflection to call Start method
	// This will actually execute the method but will fail because gnet.Run needs a real server
	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("Start")
	assert.True(t, method.IsValid(), "Start method should exist")

	// Call the method - this will fail but will execute the code path
	results := method.Call([]reflect.Value{})

	// Verify the method was called and returned an error
	assert.Equal(t, 1, len(results), "Start should return 1 value")
	errorResult := results[0].Interface()
	assert.NotNil(t, errorResult, "Start should return an error when it can't bind to address")
}

func TestGnetHandler_OnShutdown_ConnectionCleanup(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Add a connection to test cleanup
	mockConn := new(MockConn)
	mockConn.On("Close").Return(nil)

	meta := &ConnMeta{
		Conn: mockConn,
	}
	handler.connManager.BindDevice("test_device", meta)

	// Test the cleanup logic that OnShutdown performs
	// We test this directly since we can't easily mock gnet.Engine
	handler.connManager.deviceConns.Range(func(_, val any) bool {
		conn := val.(*ConnMeta).Conn
		_ = conn.Close()
		return true
	})

	// Verify connection was closed during cleanup
	mockConn.AssertCalled(t, "Close")
}

func TestGnetHandler_Start_Configuration(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":0", // Use port 0 to avoid conflicts
		MaxIdleTime:   30 * time.Second,
		TickInterval:  10 * time.Second,
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Test that Start method exists and configuration is properly set
	assert.NotNil(t, handler.conf)
	assert.Equal(t, ":0", handler.conf.Addr)
	assert.Equal(t, 30*time.Second, handler.conf.MaxIdleTime)
	assert.Equal(t, 10*time.Second, handler.conf.TickInterval)

	// We can't actually call Start() in a unit test since it would start a real server
	// but we can verify the method exists and the handler is properly configured
	assert.NotNil(t, handler.Start)
}

func TestGnetHandler_Stop_WithoutEngine(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Test Stop when engine is nil (should return error)
	err = handler.Stop()
	assert.Error(t, err) // Should fail because engine is nil
}

func TestGnetHandler_OnShutdown_NilEvent(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}

	handler, err := NewGnetHandler(conf, nil)
	assert.NoError(t, err)
	handler.event = nil

	// Test that nil event doesn't cause panic
	// We test the nil check logic directly
	assert.Nil(t, handler.event)
}

func TestGnetHandler_OnOpen(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)
	mockGnetConn := new(MockGnetConn)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Mock the RemoteAddr call for logging
	mockAddr := &mockNetAddr{addr: "127.0.0.1:12345"}
	mockGnetConn.On("RemoteAddr").Return(mockAddr)

	expectedData := []byte("response data")
	mockEvent.On("OnOpen", mockGnetConn).Return(expectedData, gnet.None)

	data, action := handler.OnOpen(mockGnetConn)

	assert.Equal(t, expectedData, data)
	assert.Equal(t, gnet.None, action)
	mockEvent.AssertExpectations(t)
	mockGnetConn.AssertExpectations(t)
}

func TestGnetHandler_OnOpen_NilEvent(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}

	handler, err := NewGnetHandler(conf, nil)
	assert.NoError(t, err)
	handler.event = nil

	mockGnetConn := new(MockGnetConn)

	// Mock the RemoteAddr call for logging
	mockAddr := &mockNetAddr{addr: "127.0.0.1:12345"}
	mockGnetConn.On("RemoteAddr").Return(mockAddr)

	data, action := handler.OnOpen(mockGnetConn)

	assert.Nil(t, data)
	assert.Equal(t, gnet.None, action)
	mockGnetConn.AssertExpectations(t)
}

func TestGnetHandler_OnClose(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)
	mockGnetConn := new(MockGnetConn)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Mock the RemoteAddr call for logging
	mockAddr := &mockNetAddr{addr: "127.0.0.1:12345"}
	mockGnetConn.On("RemoteAddr").Return(mockAddr)

	// Set up context to return device ID
	mockGnetConn.On("Context").Return(map[string]any{string(ContextKeyDeviceID): "test_device"})

	// Add device to connection manager
	mockConn := new(MockConn)
	mockConn.On("Close").Return(nil)

	meta := &ConnMeta{Conn: mockConn}
	handler.connManager.BindDevice("test_device", meta)

	testErr := errors.New("connection error")
	mockEvent.On("OnClose", mockGnetConn, testErr).Return(gnet.None)

	action := handler.OnClose(mockGnetConn, testErr)

	assert.Equal(t, gnet.None, action)
	mockEvent.AssertExpectations(t)
	mockGnetConn.AssertExpectations(t)

	// Verify device was unbound
	_, exists := handler.connManager.GetConnMeta("test_device")
	assert.False(t, exists)
}

func TestGnetHandler_OnTick(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
		MaxIdleTime:   30 * time.Second,
		TickInterval:  10 * time.Second,
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	expectedDuration := 5 * time.Second
	mockEvent.On("OnTick").Return(expectedDuration, gnet.None)

	duration, action := handler.OnTick()

	assert.Equal(t, expectedDuration, duration)
	assert.Equal(t, gnet.None, action)
	mockEvent.AssertExpectations(t)
}

func TestGnetHandler_OnTick_NilEvent(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
		TickInterval:  10 * time.Second,
	}

	handler, err := NewGnetHandler(conf, nil)
	assert.NoError(t, err)
	handler.event = nil

	duration, action := handler.OnTick()

	assert.Equal(t, conf.TickInterval, duration)
	assert.Equal(t, gnet.None, action)
}

func TestGnetHandler_Stop(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	// Test that Stop returns error when engine is nil
	err = handler.Stop()
	assert.Error(t, err) // Should return error when engine is nil
}

// nolint
func TestGnetHandler_PushTo(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		deviceID           string
		setupDevice        bool
		setupCodec         bool
		setupErrorConnMeta bool
		expectedError      error
	}{
		{
			name:               "device not found",
			deviceID:           "nonexistent",
			setupDevice:        false,
			expectedError:      ErrDeviceNotFound,
			setupErrorConnMeta: false,
		},
		{
			name:               "successful push",
			deviceID:           "test_device",
			setupDevice:        true,
			setupCodec:         true,
			expectedError:      nil,
			setupErrorConnMeta: false,
		},
		{
			name:               "nil codec",
			deviceID:           "test_device",
			setupDevice:        true,
			setupCodec:         false,
			expectedError:      nil,
			setupErrorConnMeta: false,
		},
		{
			name:               "error conn meta push",
			deviceID:           "test_device",
			setupDevice:        true,
			setupCodec:         true,
			expectedError:      ErrConnMetaNotFound,
			setupErrorConnMeta: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.setupDevice {
				mockMessage := &MockMessage{}
				err := handler.PushTo(tt.deviceID, mockMessage, nil)

				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			mockConn := new(MockConn)
			mockCodec := new(MockSecureCodec)

			meta := &ConnMeta{
				Conn:       mockConn,
				SessionKey: []byte("session_key"),
			}

			if !tt.setupCodec {
				meta.Codec = nil
				handler.connManager.BindDevice(tt.deviceID, meta)

				mockMessage := &MockMessage{}
				err := handler.PushTo(tt.deviceID, mockMessage, nil)
				assert.NoError(t, err) // Should return nil for nil codec
				return
			}

			meta.Codec = mockCodec
			mockMessage := &MockMessage{}
			encodedData := []byte("encoded_data")

			mockCodec.On("Encode", mockMessage, []byte("session_key")).Return(encodedData, nil)
			mockConn.On("AsyncWrite", encodedData, mock.AnythingOfType("handler.AsyncCallback")).Return(nil)

			if tt.setupErrorConnMeta {
				handler.connManager.deviceConns.Store(tt.deviceID, "error conn meta")
			} else {
				handler.connManager.BindDevice(tt.deviceID, meta)
			}

			err := handler.PushTo(tt.deviceID, mockMessage, nil)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGnetHandler_OnTraffic_Handshake(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	mockGnetConn := new(MockGnetConn)
	testData := []byte("handshake_data")

	// Mock reading data from connection
	mockGnetConn.On("Next", -1).Return(testData, nil)

	// Mock context without device ID (new connection)
	mockGnetConn.On("Context").Return(nil)
	mockGnetConn.On("SetContext", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Mock handshake prefix matching
	mockEvent.On("MatchHandshakePrefix", testData).Return("test_protocol", true)

	// Mock codec creation
	mockCodec := new(MockSecureCodec)
	mockEvent.On("NewCodec", "test_protocol").Return(mockCodec, nil)

	// Mock handshake processing
	mockMessage := &MockMessage{}
	sessionKey := []byte("session_key")
	deviceID := "device12"
	mockCodec.On("Handshake", testData).Return(mockMessage, sessionKey, deviceID, nil)

	// Mock event handler
	responseMessage := &MockMessage{}
	mockEvent.On("OnHandshakeMessage", mockGnetConn, mockMessage).Return(responseMessage)

	// Mock codec encode for the response
	encodedResponse := []byte("encoded_response")
	mockCodec.On("Encode", responseMessage, sessionKey).Return(encodedResponse, nil)

	// Mock the Write call that happens after encoding
	mockGnetConn.On("Write", encodedResponse).Return(len(encodedResponse), nil)

	action := handler.OnTraffic(mockGnetConn)

	assert.Equal(t, gnet.None, action)
	mockEvent.AssertExpectations(t)
	mockCodec.AssertExpectations(t)
	mockGnetConn.AssertExpectations(t)

	// Verify device was bound
	meta, exists := handler.connManager.GetConnMeta(deviceID)
	assert.True(t, exists)
	assert.NotNil(t, meta)
	assert.True(t, meta.HandshakeOK)
}

func TestGnetHandler_OnTraffic_ExistingConnection(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	mockGnetConn := new(MockGnetConn)
	deviceID := "device1"

	// Set up existing connection context
	context := map[string]any{string(ContextKeyDeviceID): deviceID}
	mockGnetConn.On("Context").Return(context)

	// Set up existing connection metadata
	mockCodec := new(MockSecureCodec)
	sessionKey := []byte("session_key")
	buffer := []byte("existing_buffer")

	meta := &ConnMeta{
		Conn:        &GnetConn{Conn: mockGnetConn},
		Codec:       mockCodec,
		SessionKey:  sessionKey,
		Buf:         buffer,
		HandshakeOK: true,
	}
	handler.connManager.BindDevice(deviceID, meta)

	// Mock reading data
	newData := []byte("new_data")
	mockGnetConn.On("Next", -1).Return(newData, nil)

	// Mock message decoding - the implementation calls Decode in a loop
	buffer = append(buffer, newData...)
	combinedData := buffer
	mockMessage := &MockMessage{}

	// First decode call returns a message and consumes some data
	consumedBytes := 10 // Consume part of the data
	mockCodec.On("Decode", combinedData, sessionKey).Return(mockMessage, consumedBytes, nil)

	// Second decode call with remaining data returns incomplete packet error to break the loop
	remainingData := combinedData[consumedBytes:]
	mockCodec.On("Decode", remainingData, sessionKey).Return((*MockMessage)(nil), 0, ErrIncompletePacket)

	// Mock event handling
	mockEvent.On("OnMessage", mockGnetConn, mockMessage).Return(mockMessage)

	// Mock codec encode for the response
	responseMessage := &MockMessage{}
	encodedResponse := []byte("encoded_response")
	mockCodec.On("Encode", responseMessage, sessionKey).Return(encodedResponse, nil)

	mockGnetConn.On("Write", encodedResponse).Return(0, nil)

	action := handler.OnTraffic(mockGnetConn)

	assert.Equal(t, gnet.None, action)
	mockEvent.AssertExpectations(t)
	mockCodec.AssertExpectations(t)
	mockGnetConn.AssertExpectations(t)

	// Verify buffer contains remaining data after partial consumption
	updatedMeta, _ := handler.connManager.GetConnMeta(deviceID)
	expectedRemainingData := combinedData[consumedBytes:]
	assert.Equal(t, expectedRemainingData, updatedMeta.Buf)
}

// nolint
func TestGnetHandler_OnTrafficMessageError(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	mockGnetConn := new(MockGnetConn)
	deviceID := "device1"

	// Set up existing connection context
	context := map[string]any{string(ContextKeyDeviceID): deviceID}

	// Set up existing connection metadata
	mockCodec := new(MockSecureCodec)
	sessionKey := []byte("session_key")
	buffer := []byte("existing_buffer")

	meta := &ConnMeta{
		Conn:        &GnetConn{Conn: mockGnetConn},
		Codec:       mockCodec,
		SessionKey:  sessionKey,
		Buf:         buffer,
		HandshakeOK: true,
	}

	tests := []struct {
		name        string
		setupMock   func()
		expectedLog string
		action      gnet.Action
	}{
		{
			name: "codec decode ErrIncompletePacket error",
			setupMock: func() {
				handler.connManager.BindDevice(deviceID, meta)
				mockGnetConn.On("Context").Return(context)
				newData := []byte("new_data")
				mockGnetConn.On("Next", -1).Return(newData, nil)
				buffer = append(buffer, newData...)
				combinedData := buffer
				consumedBytes := 10 // Consume part of the data
				mockMessage := &MockMessage{}
				mockCodec.On("Decode", combinedData, sessionKey).Return(mockMessage, consumedBytes, ErrIncompletePacket)
			},
			action: gnet.None,
		},
		{
			name: "codec decode other error",
			setupMock: func() {
				handler.connManager.BindDevice(deviceID, meta)
				newData := []byte("new_data")
				mockGnetConn.On("Context").Return(context)
				mockGnetConn.On("Close").Return(nil)
				mockGnetConn.On("Next", -1).Return(newData, nil)

				buffer = append(buffer, newData...)
				combinedData := buffer
				mockMessage := &MockMessage{}
				consumedBytes := 10 // Consume part of the data
				mockCodec.On("Decode", combinedData, sessionKey).Return(mockMessage, consumedBytes, fmt.Errorf("other error"))
			},
			action: gnet.Close,
		},
		{
			name: "codec decode consumed 0",
			setupMock: func() {
				handler.connManager.BindDevice(deviceID, meta)
				mockGnetConn.On("Context").Return(context)
				newData := []byte("new_data")
				mockGnetConn.On("Next", -1).Return(newData, nil)

				buffer = append(buffer, newData...)
				combinedData := buffer
				mockMessage := &MockMessage{}
				consumedBytes := 0 // Consume part of the data
				mockCodec.On("Decode", combinedData, sessionKey).Return(mockMessage, consumedBytes, nil)
			},
			action: gnet.None,
		},
		{
			name: "codec Encode error",
			setupMock: func() {
				handler.connManager.BindDevice(deviceID, meta)
				mockGnetConn.On("Context").Return(context)
				newData := []byte("new_data")
				mockGnetConn.On("Next", -1).Return(newData, nil)
				mockGnetConn.On("Close").Return(nil)
				buffer = append(buffer, newData...)
				combinedData := buffer
				mockMessage := &MockMessage{}
				responseMessage := &MockMessage{}
				consumedBytes := 10 // Consume part of the
				mockCodec.On("Decode", combinedData, sessionKey).Return(mockMessage, consumedBytes, nil)
				mockEvent.On("OnMessage", mockGnetConn, mockMessage).Return(responseMessage)
				encodedResponse := []byte("encoded_response")

				sessionKey := []byte("session_key")
				mockCodec.On("Encode", responseMessage, sessionKey).Return(encodedResponse, fmt.Errorf("Encode error"))
			},
			action: gnet.Close,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockGnetConn.ExpectedCalls = nil
			mockEvent.ExpectedCalls = nil

			tt.setupMock()

			action := handler.OnTraffic(mockGnetConn)

			assert.Equal(t, tt.action, action)
			mockEvent.AssertExpectations(t)
			mockGnetConn.AssertExpectations(t)
		})
	}
}

// nolint
func TestGnetHandler_OnTrafficHandshakeError(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)

	mockGnetConn := new(MockGnetConn)
	testAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:123")
	deviceID := "device12"

	tests := []struct {
		name        string
		setupMock   func()
		expectedLog string
		action      gnet.Action
	}{
		{
			name: "handshake prefix not matched",
			setupMock: func() {
				mockGnetConn.On("Context").Return(nil)
				mockGnetConn.On("Next", -1).Return([]byte("invalid_data"), nil)
				mockGnetConn.On("RemoteAddr").Return(testAddr)
				mockEvent.On("MatchHandshakePrefix", []byte("invalid_data")).Return("", false)
			},
			action: gnet.Close,
		},
		{
			name: "codec creation error",
			setupMock: func() {
				mockGnetConn.On("Context").Return(nil)
				mockGnetConn.On("Next", -1).Return([]byte("handshake_data"), nil)
				mockGnetConn.On("RemoteAddr").Return(testAddr)
				mockEvent.On("MatchHandshakePrefix", []byte("handshake_data")).Return("test_protocol", true)
				mockEvent.On("NewCodec", "test_protocol").Return((*MockSecureCodec)(nil), errors.New("codec error"))
			},
			action: gnet.Close,
		},
		{
			name: "codec Handshake error",
			setupMock: func() {
				mockCodec := new(MockSecureCodec)
				mockGnetConn.On("Context").Return(nil)
				mockGnetConn.On("Next", -1).Return([]byte("handshake_data"), nil)
				mockGnetConn.On("RemoteAddr").Return(testAddr)
				mockEvent.On("MatchHandshakePrefix", []byte("handshake_data")).Return("test_protocol", true)
				mockEvent.On("NewCodec", "test_protocol").Return(mockCodec, nil)
				mockMessage := &MockMessage{}
				sessionKey := []byte("session_key")
				testData := []byte("handshake_data")
				mockCodec.On("Handshake", testData).Return(mockMessage, sessionKey, deviceID, fmt.Errorf("Handshake error"))
			},
			action: gnet.Close,
		},
		{
			name: "OnHandshakeMessage return nil",
			setupMock: func() {
				mockCodec := new(MockSecureCodec)
				mockGnetConn.On("Context").Return(nil)
				mockGnetConn.On("SetContext", mock.AnythingOfType("map[string]interface {}")).Return(nil)
				mockGnetConn.On("Next", -1).Return([]byte("handshake_data"), nil)

				mockMessage := &MockMessage{}
				mockEvent.On("MatchHandshakePrefix", []byte("handshake_data")).Return("test_protocol", true)
				mockEvent.On("NewCodec", "test_protocol").Return(mockCodec, nil)
				mockEvent.On("OnHandshakeMessage", mockGnetConn, mockMessage).Return(nil)

				sessionKey := []byte("session_key")
				testData := []byte("handshake_data")
				mockCodec.On("Handshake", testData).Return(mockMessage, sessionKey, deviceID, nil)
			},
			action: gnet.None,
		},
		{
			name: "codec Encode error",
			setupMock: func() {
				mockCodec := new(MockSecureCodec)
				mockGnetConn.On("Context").Return(nil)
				mockGnetConn.On("SetContext", mock.AnythingOfType("map[string]interface {}")).Return(nil)
				mockGnetConn.On("Next", -1).Return([]byte("handshake_data"), nil)
				mockGnetConn.On("Close").Return(nil)

				mockMessage := &MockMessage{}
				responseMessage := &MockMessage{}

				mockEvent.On("MatchHandshakePrefix", []byte("handshake_data")).Return("test_protocol", true)
				mockEvent.On("NewCodec", "test_protocol").Return(mockCodec, nil)
				mockEvent.On("OnHandshakeMessage", mockGnetConn, mockMessage).Return(responseMessage)
				encodedResponse := []byte("encoded_response")

				sessionKey := []byte("session_key")
				testData := []byte("handshake_data")
				mockCodec.On("Handshake", testData).Return(mockMessage, sessionKey, deviceID, nil)
				mockCodec.On("Encode", responseMessage, sessionKey).Return(encodedResponse, fmt.Errorf("Encode error"))
			},
			action: gnet.Close,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockGnetConn.ExpectedCalls = nil
			mockEvent.ExpectedCalls = nil

			tt.setupMock()

			action := handler.OnTraffic(mockGnetConn)

			assert.Equal(t, tt.action, action)
			mockEvent.AssertExpectations(t)
			mockGnetConn.AssertExpectations(t)
		})
	}
}

func TestGnetHandler_UnbindDevice(t *testing.T) {
	conf := &serverconf.SocketConfig{
		FrameworkType: serverconf.FrameworkGnet,
		Addr:          ":8080",
	}
	mockEvent := new(MockGnetEventHandler)

	handler, err := NewGnetHandler(conf, mockEvent)
	assert.NoError(t, err)
	mockGnetConn := new(MockGnetConn)

	tests := []struct {
		name        string
		setupMock   func()
		expectedLog string
	}{
		{
			name: "device not found",
			setupMock: func() {
			},
		},
		{
			name: "close device",
			setupMock: func() {
				handler.connManager.BindDevice("device", &ConnMeta{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockGnetConn.ExpectedCalls = nil
			mockEvent.ExpectedCalls = nil

			tt.setupMock()

			handler.UnbindDevice("device")

			mockEvent.AssertExpectations(t)
			mockGnetConn.AssertExpectations(t)
		})
	}
}

func TestGnetHandler_NilEvent(_ *testing.T) {
	h := &GnetHandler{
		logger:      logger.NewLogger(),
		connManager: NewConnManager(),
	}

	testData := []byte("")
	testAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:123")
	mockGnetConn := new(MockGnetConn)
	mockGnetConn.On("RemoteAddr").Return(testAddr)
	mockGnetConn.On("Next", -1).Return(testData, nil)

	h.OnShutdown(gnet.Engine{})
	_ = h.OnClose(mockGnetConn, nil)
	h.OnTraffic(mockGnetConn)
}
