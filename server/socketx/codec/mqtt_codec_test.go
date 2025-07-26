package codec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMQTTCodec_ProtocolName(t *testing.T) {
	codec := &MQTTCodec{}
	assert.Equal(t, "mqtt", codec.ProtocolName())
}

func TestMQTTCodec_IsSecure(t *testing.T) {
	codec := &MQTTCodec{}
	assert.True(t, codec.IsSecure())
}

func TestMQTTCodec_Split(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name            string
		buffer          []byte
		expectedPackets [][]byte
		expectedRest    []byte
		expectError     bool
	}{
		{
			name:            "empty buffer",
			buffer:          []byte{},
			expectedPackets: nil,
			expectedRest:    []byte{},
		},
		{
			name:            "buffer too short",
			buffer:          []byte{0x10},
			expectedPackets: nil,
			expectedRest:    []byte{0x10},
		},
		{
			name:            "incomplete packet",
			buffer:          []byte{0x10, 0x0A, 0x00, 0x04},
			expectedPackets: nil,
			expectedRest:    []byte{0x10, 0x0A, 0x00, 0x04},
		},
		{
			name:   "complete single packet",
			buffer: []byte{0x10, 0x02, 0x00, 0x00}, // CONNECT packet with 2 bytes remaining
			expectedPackets: [][]byte{
				{0x10, 0x02, 0x00, 0x00},
			},
			expectedRest: []byte{},
		},
		{
			name:   "packet with remaining data",
			buffer: []byte{0x10, 0x02, 0x00, 0x00, 0x20, 0x01, 0xFF}, // CONNECT + extra data
			expectedPackets: [][]byte{
				{0x10, 0x02, 0x00, 0x00},
			},
			expectedRest: []byte{0x20, 0x01, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packets, rest, err := codec.Split(tt.buffer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPackets, packets)
				assert.Equal(t, tt.expectedRest, rest)
			}
		})
	}
}

// nolint
func TestMQTTCodec_Handshake(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name           string
		data           []byte
		expectedMsg    *MQTTMessage
		expectedKey    []byte
		expectedDevice string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "empty data",
			data:          []byte{},
			expectError:   true,
			errorContains: "handshake packet invalid or incomplete",
		},
		{
			name:          "invalid packet type",
			data:          []byte{0x20, 0x02, 0x00, 0x00}, // CONNACK instead of CONNECT
			expectError:   true,
			errorContains: "expected CONNECT packet, got 2",
		},
		{
			name:          "incomplete protocol name",
			data:          []byte{0x10, 0x02, 0x00},
			expectError:   true,
			errorContains: "handshake packet invalid or incomplete", // Split fails first
		},
		{
			name:           "valid CONNECT packet",
			data:           BuildMQTTConnectPacket("test_client"),
			expectedDevice: "test_client",
			expectedKey:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, key, deviceID, err := codec.Handshake(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, msg)
				assert.Nil(t, key)
				assert.Empty(t, deviceID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.Equal(t, tt.expectedKey, key)
				assert.Equal(t, tt.expectedDevice, deviceID)

				// Verify message content
				mqttMsg, ok := msg.(*MQTTMessage)
				assert.True(t, ok)
				assert.NotNil(t, mqttMsg.GetPayload())

				header := mqttMsg.GetHeader()
				assert.Equal(t, "MQTT", header["protocol"])
				assert.Equal(t, "4", header["version"])
				assert.Equal(t, tt.expectedDevice, header["clientID"])
				assert.Equal(t, "60", header["keepAlive"])
				assert.Equal(t, "true", header["clean"])
			}
		})
	}
}

// nolint
func TestMQTTCodec_Encode(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name        string
		message     Message
		key         []byte
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing topic",
			message: &MQTTMessage{
				Payload: []byte("test payload"),
				Header:  map[string]string{},
			},
			expectError: true,
			errorMsg:    "missing topic",
		},
		{
			name: "basic publish message",
			message: &MQTTMessage{
				Payload: []byte("hello world"),
				Header: map[string]string{
					"topic": "test/topic",
				},
			},
		},
		{
			name: "publish with QoS 1",
			message: &MQTTMessage{
				Payload: []byte("qos1 message"),
				Header: map[string]string{
					"topic": "test/qos1",
					"qos":   "1",
				},
			},
		},
		{
			name: "publish with retain flag",
			message: &MQTTMessage{
				Payload: []byte("retained message"),
				Header: map[string]string{
					"topic":  "test/retain",
					"retain": "1",
				},
			},
		},
		{
			name: "publish with QoS 1 and retain",
			message: &MQTTMessage{
				Payload: []byte("qos1 retained"),
				Header: map[string]string{
					"topic":  "test/both",
					"qos":    "1",
					"retain": "1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := codec.Encode(tt.message, tt.key)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, encoded)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encoded)
				assert.True(t, len(encoded) > 0)

				// Verify the packet structure
				assert.Equal(t, byte(0x30), encoded[0]&0xF0) // PUBLISH packet type

				// Check flags based on message headers
				header := tt.message.GetHeader()
				expectedFlags := byte(0x30) // Base PUBLISH packet

				if retain, ok := header["retain"]; ok && retain == "1" {
					expectedFlags |= 0x01
				}
				if qos, ok := header["qos"]; ok && qos == "1" {
					expectedFlags |= 0x02
				}

				assert.Equal(t, expectedFlags, encoded[0])
			}
		})
	}
}

// nolint
func TestMQTTCodec_Decode(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name             string
		data             []byte
		key              []byte
		expectedMsg      *MQTTMessage
		expectedConsumed int
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "empty data",
			data:             []byte{},
			expectedMsg:      nil,
			expectedConsumed: 0,
		},
		{
			name:             "data too short",
			data:             []byte{0x30},
			expectedMsg:      nil,
			expectedConsumed: 0,
		},
		{
			name:        "non-publish packet",
			data:        []byte{0x10, 0x02, 0x00, 0x00}, // CONNECT packet
			expectError: true,
			errorMsg:    "unsupported packet type",
		},
		{
			name:             "incomplete publish packet",
			data:             []byte{0x30, 0x10, 0x00, 0x04}, // Incomplete
			expectedMsg:      nil,
			expectedConsumed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, consumed, err := codec.Decode(tt.data, tt.key)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, msg)
				assert.Equal(t, 0, consumed)
			} else {
				if tt.expectedMsg == nil {
					assert.NoError(t, err)
					assert.Nil(t, msg)
					assert.Equal(t, tt.expectedConsumed, consumed)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, msg)
					assert.Equal(t, tt.expectedConsumed, consumed)

					mqttMsg, ok := msg.(*MQTTMessage)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedMsg.Payload, mqttMsg.Payload)
					assert.Equal(t, tt.expectedMsg.Header, mqttMsg.Header)
				}
			}
		})
	}
}

func TestMQTTCodec_EncodeDecodeRoundTrip(t *testing.T) {
	codec := &MQTTCodec{}

	originalMsg := &MQTTMessage{
		Payload: []byte("round trip test"),
		Header: map[string]string{
			"topic":  "test/roundtrip",
			"qos":    "1",
			"retain": "1",
		},
	}

	// Encode the message
	encoded, err := codec.Encode(originalMsg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, encoded)

	// Decode the encoded message
	decodedMsg, consumed, err := codec.Decode(encoded, nil)
	assert.NoError(t, err)
	assert.NotNil(t, decodedMsg)
	assert.Equal(t, len(encoded), consumed)

	// Verify the decoded message matches the original
	mqttMsg, ok := decodedMsg.(*MQTTMessage)
	assert.True(t, ok)
	assert.Equal(t, originalMsg.Payload, mqttMsg.Payload)
	assert.Equal(t, originalMsg.Header["topic"], mqttMsg.Header["topic"])
	assert.Equal(t, "true", mqttMsg.Header["retain"]) // Should be "true" string
	assert.Equal(t, "1", mqttMsg.Header["qos"])
}

func TestMQTTMessage(t *testing.T) {
	payload := []byte("test payload")
	header := map[string]string{
		"topic": "test/topic",
		"qos":   "1",
	}

	msg := &MQTTMessage{
		Payload: payload,
		Header:  header,
	}

	// Test GetPayload
	assert.Equal(t, payload, msg.GetPayload())

	// Test GetHeader
	assert.Equal(t, header, msg.GetHeader())

	// Verify it implements Message interface
	var _ Message = msg
}

func TestEncodeRemainingLength(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
		length   int
	}{
		{
			name:     "zero length",
			length:   0,
			expected: []byte{0x00},
		},
		{
			name:     "small length",
			length:   127,
			expected: []byte{0x7F},
		},
		{
			name:     "medium length",
			length:   128,
			expected: []byte{0x80, 0x01},
		},
		{
			name:     "large length",
			length:   16383,
			expected: []byte{0xFF, 0x7F},
		},
		{
			name:     "very large length",
			length:   16384,
			expected: []byte{0x80, 0x80, 0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeRemainingLength(tt.length)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeRemainingLength(t *testing.T) {
	tests := []struct {
		name           string
		buffer         []byte
		expectedValue  int
		expectedLength int
	}{
		{
			name:           "empty buffer",
			buffer:         []byte{},
			expectedValue:  0,
			expectedLength: 0,
		},
		{
			name:           "single byte",
			buffer:         []byte{0x7F},
			expectedValue:  127,
			expectedLength: 1,
		},
		{
			name:           "two bytes",
			buffer:         []byte{0x80, 0x01},
			expectedValue:  128,
			expectedLength: 2,
		},
		{
			name:           "three bytes",
			buffer:         []byte{0xFF, 0x7F},
			expectedValue:  16383,
			expectedLength: 2,
		},
		{
			name:           "incomplete sequence",
			buffer:         []byte{0x80},
			expectedValue:  0,
			expectedLength: 0,
		},
		{
			name:           "with extra data",
			buffer:         []byte{0x7F, 0xFF, 0xFF},
			expectedValue:  127,
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, length := DecodeRemainingLength(tt.buffer)
			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedLength, length)
		})
	}
}

func TestEncodeDecodeRemainingLengthRoundTrip(t *testing.T) {
	testLengths := []int{0, 1, 127, 128, 255, 16383, 16384, 2097151}

	for _, length := range testLengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			encoded := EncodeRemainingLength(length)
			decoded, consumedBytes := DecodeRemainingLength(encoded)

			assert.Equal(t, length, decoded)
			assert.Equal(t, len(encoded), consumedBytes)
		})
	}
}

func TestBuildMQTTConnectPacket(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
	}{
		{
			name:     "empty client ID",
			clientID: "",
		},
		{
			name:     "short client ID",
			clientID: "test",
		},
		{
			name:     "long client ID",
			clientID: "very_long_client_identifier_for_testing",
		},
		{
			name:     "client ID with special characters",
			clientID: "client-123_test.device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := BuildMQTTConnectPacket(tt.clientID)

			// Verify packet structure
			assert.True(t, len(packet) > 0)
			assert.Equal(t, byte(0x10), packet[0]) // CONNECT packet type

			// Test that the packet can be parsed by handshake
			codec := &MQTTCodec{}
			msg, key, deviceID, err := codec.Handshake(packet)

			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Nil(t, key)
			assert.Equal(t, tt.clientID, deviceID)

			// Verify message header
			header := msg.GetHeader()
			assert.Equal(t, "MQTT", header["protocol"])
			assert.Equal(t, "4", header["version"])
			assert.Equal(t, tt.clientID, header["clientID"])
			assert.Equal(t, "60", header["keepAlive"])
			assert.Equal(t, "true", header["clean"])
		})
	}
}
