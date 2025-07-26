package handler

import (
	"testing"
	"time"

	"github.com/hewen/mastiff-go/server/socketx/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecureCodec is a mock implementation of codec.SecureCodec.
type MockSecureCodec struct {
	mock.Mock
}

func (m *MockSecureCodec) Split(buffer []byte) ([][]byte, []byte, error) {
	args := m.Called(buffer)
	return args.Get(0).([][]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockSecureCodec) Handshake(data []byte) (codec.Message, []byte, string, error) {
	args := m.Called(data)
	return args.Get(0).(codec.Message), args.Get(1).([]byte), args.String(2), args.Error(3)
}

func (m *MockSecureCodec) Encode(msg codec.Message, key []byte) ([]byte, error) {
	args := m.Called(msg, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSecureCodec) Decode(data []byte, key []byte) (codec.Message, int, error) {
	args := m.Called(data, key)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}

	return args.Get(0).(codec.Message), args.Int(1), args.Error(2)
}

func (m *MockSecureCodec) IsSecure() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSecureCodec) ProtocolName() string {
	args := m.Called()
	return args.String(0)
}

func TestNewConnManager(t *testing.T) {
	cm := NewConnManager()
	assert.NotNil(t, cm)
}

func TestConnManager_BindDevice(t *testing.T) {
	cm := NewConnManager()
	mockConn := new(MockConn)
	mockCodec := new(MockSecureCodec)

	deviceID := "device1234"
	meta := &ConnMeta{
		Conn:        mockConn,
		Codec:       mockCodec,
		SessionKey:  []byte("session_key"),
		HandshakeOK: true,
	}

	// Test binding a device
	cm.BindDevice(deviceID, meta)

	// Verify the device was bound
	retrievedMeta, exists := cm.GetConnMeta(deviceID)
	assert.True(t, exists)
	assert.Equal(t, meta.Conn, retrievedMeta.Conn)
	assert.Equal(t, meta.Codec, retrievedMeta.Codec)
	assert.Equal(t, meta.SessionKey, retrievedMeta.SessionKey)
	assert.Equal(t, meta.HandshakeOK, retrievedMeta.HandshakeOK)
	assert.True(t, retrievedMeta.LastActive > 0)
}

func TestConnManager_GetConnMeta(t *testing.T) {
	cm := NewConnManager()
	mockConn := new(MockConn)
	mockCodec := new(MockSecureCodec)

	tests := []struct {
		name           string
		deviceID       string
		setupDevice    bool
		expectedExists bool
	}{
		{
			name:           "get existing device",
			deviceID:       "device123",
			setupDevice:    true,
			expectedExists: true,
		},
		{
			name:           "get non-existing device",
			deviceID:       "nonexistent",
			setupDevice:    false,
			expectedExists: false,
		},
		{
			name:           "get with empty device ID",
			deviceID:       "",
			setupDevice:    false,
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupDevice {
				meta := &ConnMeta{
					Conn:        mockConn,
					Codec:       mockCodec,
					SessionKey:  []byte("session_key"),
					HandshakeOK: true,
				}
				cm.BindDevice(tt.deviceID, meta)
			}

			retrievedMeta, exists := cm.GetConnMeta(tt.deviceID)
			assert.Equal(t, tt.expectedExists, exists)
			if tt.expectedExists {
				assert.NotNil(t, retrievedMeta)
				assert.Equal(t, mockConn, retrievedMeta.Conn)
			} else {
				assert.Nil(t, retrievedMeta)
			}
		})
	}
}

func TestConnManager_UnbindDevice(t *testing.T) {
	cm := NewConnManager()
	mockConn := new(MockConn)
	mockCodec := new(MockSecureCodec)

	mockConn.On("Close").Return(nil)

	deviceID := "device123"
	meta := &ConnMeta{
		Conn:        mockConn,
		Codec:       mockCodec,
		SessionKey:  []byte("session_key"),
		HandshakeOK: true,
	}

	// Bind the device first
	cm.BindDevice(deviceID, meta)
	_, exists := cm.GetConnMeta(deviceID)
	assert.True(t, exists)

	// Unbind the device
	cm.UnbindDevice(deviceID)

	// Verify the device was unbound
	_, exists = cm.GetConnMeta(deviceID)
	assert.False(t, exists)

	// Test unbinding non-existing device (should not panic)
	cm.UnbindDevice("nonexistent")
}

func TestConnManager_UpdateActivity(t *testing.T) {
	cm := NewConnManager()
	mockConn := new(MockConn)
	mockCodec := new(MockSecureCodec)

	deviceID := "device123"
	meta := &ConnMeta{
		Conn:        mockConn,
		Codec:       mockCodec,
		SessionKey:  []byte("session_key"),
		HandshakeOK: true,
		LastActive:  time.Now().Unix() - 100, // Set to 100 seconds ago
	}

	// Bind the device
	cm.BindDevice(deviceID, meta)

	// Store the initial time for comparison
	initialTime := time.Now().Unix() - 100

	// Wait a moment to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update activity
	cm.UpdateActivity(deviceID)

	// Verify last active time was updated
	updatedMeta, exists := cm.GetConnMeta(deviceID)
	assert.True(t, exists)
	assert.True(t, updatedMeta.LastActive > initialTime, "LastActive should be updated to a more recent time")

	// Test updating activity for non-existing device (should not panic)
	cm.UpdateActivity("nonexistent")
}

func TestConnManager_CleanupInactive(t *testing.T) {
	cm := NewConnManager()

	// Create mock connections
	activeConn := new(MockConn)
	inactiveConn := new(MockConn)

	// Set up expectations for closing inactive connection
	inactiveConn.On("Close").Return(nil)

	mockCodec := new(MockSecureCodec)

	now := time.Now().Unix()
	maxIdleSeconds := int64(60) // 1 minute

	// Create active device (recently active) - manually set LastActive
	activeMeta := &ConnMeta{
		Conn:       activeConn,
		Codec:      mockCodec,
		LastActive: now - 30, // 30 seconds ago (within limit)
	}
	// Store directly without using BindDevice to avoid LastActive being overwritten
	cm.deviceConns.Store("active_device", activeMeta)

	// Create inactive device (old activity) - manually set LastActive
	inactiveMeta := &ConnMeta{
		Conn:       inactiveConn,
		Codec:      mockCodec,
		LastActive: now - 120, // 2 minutes ago (beyond limit)
	}
	// Store directly without using BindDevice to avoid LastActive being overwritten
	cm.deviceConns.Store("inactive_device", inactiveMeta)

	// Perform cleanup
	cm.CleanupInactive(maxIdleSeconds)

	// Verify active device still exists
	_, exists := cm.GetConnMeta("active_device")
	assert.True(t, exists)

	// Verify inactive device was removed
	_, exists = cm.GetConnMeta("inactive_device")
	assert.False(t, exists)

	// Verify inactive connection was closed
	inactiveConn.AssertCalled(t, "Close")
	activeConn.AssertNotCalled(t, "Close")
}

func TestConnMeta_Fields(t *testing.T) {
	mockConn := new(MockConn)
	mockCodec := new(MockSecureCodec)
	sessionKey := []byte("test_session_key")
	buf := []byte("test_buffer")

	meta := &ConnMeta{
		Conn:        mockConn,
		Codec:       mockCodec,
		LastActive:  time.Now().Unix(),
		SessionKey:  sessionKey,
		Buf:         buf,
		HandshakeOK: true,
	}

	assert.Equal(t, mockConn, meta.Conn)
	assert.Equal(t, mockCodec, meta.Codec)
	assert.True(t, meta.LastActive > 0)
	assert.Equal(t, sessionKey, meta.SessionKey)
	assert.Equal(t, buf, meta.Buf)
	assert.True(t, meta.HandshakeOK)
}
