package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSecureCodec struct {
	mock.Mock
}

func (m *MockSecureCodec) Split(buffer []byte) ([][]byte, []byte, error) {
	args := m.Called(buffer)
	return args.Get(0).([][]byte), args.Get(1).([]byte), args.Error(2)
}

func (m *MockSecureCodec) Handshake(data []byte) (Message, []byte, string, error) {
	args := m.Called(data)
	return args.Get(0).(Message), args.Get(1).([]byte), args.String(2), args.Error(3)
}

func (m *MockSecureCodec) Encode(msg Message, key []byte) ([]byte, error) {
	args := m.Called(msg, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSecureCodec) Decode(data []byte, key []byte) (Message, int, error) {
	args := m.Called(data, key)
	return args.Get(0).(Message), args.Int(1), args.Error(2)
}

func (m *MockSecureCodec) IsSecure() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSecureCodec) ProtocolName() string {
	args := m.Called()
	return args.String(0)
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

func TestSecureCodecInterface(t *testing.T) {
	mockCodec := new(MockSecureCodec)

	// Verify that MockSecureCodec implements SecureCodec interface
	var _ SecureCodec = mockCodec

	// Test Split method
	testBuffer := []byte("test data")
	expectedPackets := [][]byte{[]byte("packet1"), []byte("packet2")}
	expectedRemaining := []byte("remaining")

	mockCodec.On("Split", testBuffer).Return(expectedPackets, expectedRemaining, nil)

	packets, remaining, err := mockCodec.Split(testBuffer)
	assert.NoError(t, err)
	assert.Equal(t, expectedPackets, packets)
	assert.Equal(t, expectedRemaining, remaining)

	// Test Handshake method
	handshakeData := []byte("handshake")
	mockMessage := new(MockMessage)
	sessionKey := []byte("session_key")
	deviceID := "device123"

	mockCodec.On("Handshake", handshakeData).Return(mockMessage, sessionKey, deviceID, nil)

	msg, key, id, err := mockCodec.Handshake(handshakeData)
	assert.NoError(t, err)
	assert.Equal(t, mockMessage, msg)
	assert.Equal(t, sessionKey, key)
	assert.Equal(t, deviceID, id)

	// Test Encode method
	encodedData := []byte("encoded")
	mockCodec.On("Encode", mockMessage, sessionKey).Return(encodedData, nil)

	encoded, err := mockCodec.Encode(mockMessage, sessionKey)
	assert.NoError(t, err)
	assert.Equal(t, encodedData, encoded)

	// Test Decode method
	decodeData := []byte("decode_data")
	consumedBytes := 10
	mockCodec.On("Decode", decodeData, sessionKey).Return(mockMessage, consumedBytes, nil)

	decodedMsg, consumed, err := mockCodec.Decode(decodeData, sessionKey)
	assert.NoError(t, err)
	assert.Equal(t, mockMessage, decodedMsg)
	assert.Equal(t, consumedBytes, consumed)

	// Test IsSecure method
	mockCodec.On("IsSecure").Return(true)

	isSecure := mockCodec.IsSecure()
	assert.True(t, isSecure)

	// Test ProtocolName method
	protocolName := "test_protocol"
	mockCodec.On("ProtocolName").Return(protocolName)

	name := mockCodec.ProtocolName()
	assert.Equal(t, protocolName, name)

	mockCodec.AssertExpectations(t)
}

func TestMessageInterface(t *testing.T) {
	mockMessage := new(MockMessage)

	// Verify that MockMessage implements Message interface
	var _ Message = mockMessage

	// Test GetPayload method
	expectedPayload := []byte("test payload")
	mockMessage.On("GetPayload").Return(expectedPayload)

	payload := mockMessage.GetPayload()
	assert.Equal(t, expectedPayload, payload)

	// Test GetHeader method
	expectedHeader := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	mockMessage.On("GetHeader").Return(expectedHeader)

	header := mockMessage.GetHeader()
	assert.Equal(t, expectedHeader, header)

	mockMessage.AssertExpectations(t)
}

func TestSecureCodecInterfaceMethods(t *testing.T) {
	// Test that all required methods are present in the interface
	mockCodec := new(MockSecureCodec)
	mockMessage := new(MockMessage)

	// Test method signatures by calling them with proper mock values
	mockCodec.On("Split", []byte(nil)).Return([][]byte{}, []byte{}, nil)
	mockCodec.On("Handshake", []byte(nil)).Return(mockMessage, []byte(nil), "", nil)
	mockCodec.On("Encode", mockMessage, []byte(nil)).Return([]byte{}, nil)
	mockCodec.On("Decode", []byte(nil), []byte(nil)).Return(mockMessage, 0, nil)
	mockCodec.On("IsSecure").Return(false)
	mockCodec.On("ProtocolName").Return("")

	// Call all methods to verify signatures
	_, _, _ = mockCodec.Split(nil)
	_, _, _, _ = mockCodec.Handshake(nil)
	_, _ = mockCodec.Encode(mockMessage, nil)
	_, _, _ = mockCodec.Decode(nil, nil)
	_ = mockCodec.IsSecure()
	_ = mockCodec.ProtocolName()

	mockCodec.AssertExpectations(t)
}

func TestMessageInterfaceMethods(t *testing.T) {
	// Test that all required methods are present in the interface
	mockMessage := new(MockMessage)

	// Test method signatures
	mockMessage.On("GetPayload").Return([]byte{})
	mockMessage.On("GetHeader").Return(map[string]string{})

	// Call all methods to verify signatures
	_ = mockMessage.GetPayload()
	_ = mockMessage.GetHeader()

	mockMessage.AssertExpectations(t)
}

func TestInterfaceCompliance(t *testing.T) {
	// Test that our mock implementations satisfy the interfaces
	t.Run("SecureCodec compliance", func(t *testing.T) {
		var codec SecureCodec = new(MockSecureCodec)
		assert.NotNil(t, codec)
	})

	t.Run("Message compliance", func(t *testing.T) {
		var message Message = new(MockMessage)
		assert.NotNil(t, message)
	})
}
