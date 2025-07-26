package handler

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGnetConn is a mock implementation of gnet.Conn.
type MockGnetConn struct {
	mock.Mock
}

func (m *MockGnetConn) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGnetConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockGnetConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockGnetConn) Context() any {
	args := m.Called()
	return args.Get(0)
}

func (m *MockGnetConn) SetContext(ctx any) {
	m.Called(ctx)
}

func (m *MockGnetConn) AsyncWrite(data []byte, callback gnet.AsyncCallback) error {
	args := m.Called(data, callback)
	return args.Error(0)
}

func (m *MockGnetConn) Wake(callback gnet.AsyncCallback) error {
	args := m.Called(callback)
	return args.Error(0)
}

func (m *MockGnetConn) CloseWithCallback(callback gnet.AsyncCallback) error {
	args := m.Called(callback)
	return args.Error(0)
}

func (m *MockGnetConn) Fd() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGnetConn) Dup() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) SetReadBuffer(bytes int) error {
	args := m.Called(bytes)
	return args.Error(0)
}

func (m *MockGnetConn) SetWriteBuffer(bytes int) error {
	args := m.Called(bytes)
	return args.Error(0)
}

func (m *MockGnetConn) SetLinger(sec int) error {
	args := m.Called(sec)
	return args.Error(0)
}

func (m *MockGnetConn) SetNoDelay(noDelay bool) error {
	args := m.Called(noDelay)
	return args.Error(0)
}

func (m *MockGnetConn) SetKeepAlivePeriod(d time.Duration) error {
	args := m.Called(d)
	return args.Error(0)
}

func (m *MockGnetConn) BufferLength() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGnetConn) InboundBuffered() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGnetConn) OutboundBuffered() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGnetConn) Discard(n int) (int, error) {
	args := m.Called(n)
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) Next(n int) ([]byte, error) {
	args := m.Called(n)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockGnetConn) Peek(n int) ([]byte, error) {
	args := m.Called(n)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockGnetConn) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGnetConn) Writev(bs [][]byte) (int, error) {
	args := m.Called(bs)
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) AsyncWritev(bs [][]byte, callback gnet.AsyncCallback) error {
	args := m.Called(bs, callback)
	return args.Error(0)
}

func (m *MockGnetConn) EventLoop() gnet.EventLoop {
	args := m.Called()
	return args.Get(0).(gnet.EventLoop)
}

func (m *MockGnetConn) ReadFrom(r io.Reader) (int64, error) {
	args := m.Called(r)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGnetConn) SendTo(buf []byte, addr net.Addr) (int, error) {
	args := m.Called(buf, addr)
	return args.Int(0), args.Error(1)
}

func (m *MockGnetConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockGnetConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockGnetConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockGnetConn) SetKeepAlive(keepalive bool, idle, interval time.Duration, count int) error {
	args := m.Called(keepalive, idle, interval, count)
	return args.Error(0)
}

func (m *MockGnetConn) WriteTo(w io.Writer) (int64, error) {
	args := m.Called(w)
	return args.Get(0).(int64), args.Error(1)
}

// nolint
func TestGnetConn_AsyncWrite(t *testing.T) {
	tests := []struct {
		data          []byte
		setupMock     func(*MockGnetConn)
		expectedError error
		name          string
	}{
		{
			name: "successful async write",
			data: []byte("test data"),
			setupMock: func(m *MockGnetConn) {
				m.On("AsyncWrite", []byte("test data"), mock.AnythingOfType("gnet.AsyncCallback")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "async write with error",
			data: []byte("test data"),
			setupMock: func(m *MockGnetConn) {
				m.On("AsyncWrite", []byte("test data"), mock.AnythingOfType("gnet.AsyncCallback")).Return(errors.New("write error"))
			},
			expectedError: errors.New("write error"),
		},
		{
			name: "async write with empty data",
			data: []byte{},
			setupMock: func(m *MockGnetConn) {
				m.On("AsyncWrite", []byte{}, mock.AnythingOfType("gnet.AsyncCallback")).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGnetConn := new(MockGnetConn)
			tt.setupMock(mockGnetConn)

			gnetConn := &GnetConn{Conn: mockGnetConn}

			// Create a test callback
			var callbackCalled bool
			var callbackConn Conn
			var callbackErr error

			callback := func(c Conn, err error) error {
				callbackCalled = true
				callbackConn = c
				callbackErr = err
				return nil
			}

			err := gnetConn.AsyncWrite(tt.data, callback)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockGnetConn.AssertExpectations(t)

			// Test the callback wrapper by calling the gnet callback that was passed
			if tt.expectedError == nil {
				// Get the gnet callback that was passed to AsyncWrite
				calls := mockGnetConn.Calls
				if len(calls) > 0 {
					call := calls[0]
					if len(call.Arguments) > 1 {
						gnetCallback := call.Arguments[1].(gnet.AsyncCallback)

						// Call the gnet callback to test our wrapper
						testErr := errors.New("test callback error")
						err := gnetCallback(mockGnetConn, testErr)

						assert.NoError(t, err) // Our callback returns nil
						assert.True(t, callbackCalled)
						assert.NotNil(t, callbackConn)
						assert.Equal(t, testErr, callbackErr)

						// Verify the connection passed to callback is wrapped
						_, ok := callbackConn.(*GnetConn)
						assert.True(t, ok)
					}
				}
			}
		})
	}
}

func TestGnetConn_Write(t *testing.T) {
	mockGnetConn := new(MockGnetConn)
	data := []byte("test data")

	mockGnetConn.On("Write", data).Return(len(data), nil)

	gnetConn := &GnetConn{Conn: mockGnetConn}

	n, err := gnetConn.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	mockGnetConn.AssertExpectations(t)
}

func TestGnetConn_Close(t *testing.T) {
	mockGnetConn := new(MockGnetConn)

	mockGnetConn.On("Close").Return(nil)

	gnetConn := &GnetConn{Conn: mockGnetConn}

	err := gnetConn.Close()

	assert.NoError(t, err)
	mockGnetConn.AssertExpectations(t)
}

func TestGnetConn_Context(t *testing.T) {
	mockGnetConn := new(MockGnetConn)
	testContext := map[string]any{"key": "value"}

	mockGnetConn.On("Context").Return(testContext)

	gnetConn := &GnetConn{Conn: mockGnetConn}

	ctx := gnetConn.Context()

	assert.Equal(t, testContext, ctx)
	mockGnetConn.AssertExpectations(t)
}

func TestGnetConn_SetContext(t *testing.T) {
	mockGnetConn := new(MockGnetConn)
	testContext := map[string]any{"key": "value"}

	mockGnetConn.On("SetContext", testContext).Return()

	gnetConn := &GnetConn{Conn: mockGnetConn}

	gnetConn.SetContext(testContext)

	mockGnetConn.AssertExpectations(t)
}

func TestGnetConn_CallbackWrapper(t *testing.T) {
	mockGnetConn := new(MockGnetConn)

	// Set up the mock to capture the callback
	var capturedCallback gnet.AsyncCallback
	mockGnetConn.On("AsyncWrite", mock.Anything, mock.AnythingOfType("gnet.AsyncCallback")).Run(func(args mock.Arguments) {
		capturedCallback = args.Get(1).(gnet.AsyncCallback)
	}).Return(nil)

	gnetConn := &GnetConn{Conn: mockGnetConn}

	// Test callback that returns an error
	callbackError := errors.New("callback error")
	callback := func(_ Conn, _ error) error {
		return callbackError
	}

	err := gnetConn.AsyncWrite([]byte("test"), callback)
	assert.NoError(t, err)

	// Now test the captured callback
	if capturedCallback != nil {
		testError := errors.New("test error")
		err := capturedCallback(mockGnetConn, testError)
		assert.Equal(t, callbackError, err)
	}
}
