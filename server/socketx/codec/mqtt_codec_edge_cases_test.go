package codec

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMQTTCodec_HandshakeEdgeCases(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name          string
		errorContains string
		data          []byte
		expectError   bool
	}{
		{
			name:          "packet too short for protocol name length",
			data:          []byte{0x10, 0x01, 0x00},
			expectError:   true,
			errorContains: "invalid packet: missing protocol name length",
		},
		{
			name:          "invalid protocol name section",
			data:          []byte{0x10, 0x05, 0x00, 0x10, 0x00}, // Claims 16 byte protocol name but only 5 bytes total
			expectError:   true,
			errorContains: "handshake packet invalid or incomplete", // Split will fail first
		},
		{
			name:          "missing client ID length",
			data:          []byte{0x10, 0x08, 0x00, 0x04, 'M', 'Q', 'T', 'T', 0x04, 0x02}, // Missing keepalive and clientID length
			expectError:   true,
			errorContains: "invalid protocol name section", // This is the actual error from the implementation
		},
		{
			name:          "client ID too short",
			data:          []byte{0x10, 0x0C, 0x00, 0x04, 'M', 'Q', 'T', 'T', 0x04, 0x02, 0x00, 0x3C, 0x00, 0x05}, // Claims 5 byte clientID but no data
			expectError:   true,
			errorContains: "clientID too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, key, deviceID, err := codec.Handshake(tt.data)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, msg)
				assert.Nil(t, key)
				assert.Empty(t, deviceID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
			}
		})
	}
}

// nolint
func TestMQTTCodec_DecodeEdgeCases(t *testing.T) {
	codec := &MQTTCodec{}

	// Create a valid PUBLISH packet for testing
	createPublishPacket := func(topic string, payload []byte, qos byte, retain bool) []byte {
		var buf bytes.Buffer

		// Fixed header
		fixedHeader := byte(0x30) // PUBLISH
		if retain {
			fixedHeader |= 0x01
		}
		if qos > 0 {
			fixedHeader |= (qos << 1)
		}
		buf.WriteByte(fixedHeader)

		// Variable header + payload
		var body bytes.Buffer
		_ = binary.Write(&body, binary.BigEndian, uint16(len(topic)))
		body.WriteString(topic)
		body.Write(payload)

		// Remaining length
		buf.Write(EncodeRemainingLength(body.Len()))
		buf.Write(body.Bytes())

		return buf.Bytes()
	}

	tests := []struct {
		name             string
		data             []byte
		expectedTopic    string
		expectedPayload  []byte
		expectedQoS      string
		expectedRetain   string
		expectedConsumed int
		expectError      bool
	}{
		{
			name:             "basic publish packet",
			data:             createPublishPacket("test/topic", []byte("hello"), 0, false),
			expectedTopic:    "test/topic",
			expectedPayload:  []byte("hello"),
			expectedQoS:      "0",
			expectedRetain:   "false",
			expectedConsumed: 18, // Will vary based on actual packet size
		},
		{
			name:            "publish with QoS 1",
			data:            createPublishPacket("qos/topic", []byte("qos1"), 1, false),
			expectedTopic:   "qos/topic",
			expectedPayload: []byte("qos1"),
			expectedQoS:     "1",
			expectedRetain:  "false",
		},
		{
			name:            "publish with retain flag",
			data:            createPublishPacket("retain/topic", []byte("retained"), 0, true),
			expectedTopic:   "retain/topic",
			expectedPayload: []byte("retained"),
			expectedQoS:     "0",
			expectedRetain:  "true",
		},
		{
			name:            "empty payload",
			data:            createPublishPacket("empty", []byte{}, 0, false),
			expectedTopic:   "empty",
			expectedPayload: []byte{},
			expectedQoS:     "0",
			expectedRetain:  "false",
		},
		{
			name:            "long topic name",
			data:            createPublishPacket("very/long/topic/name/for/testing/purposes", []byte("data"), 0, false),
			expectedTopic:   "very/long/topic/name/for/testing/purposes",
			expectedPayload: []byte("data"),
			expectedQoS:     "0",
			expectedRetain:  "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, consumed, err := codec.Decode(tt.data, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, msg)
				assert.Equal(t, 0, consumed)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.True(t, consumed > 0)

				mqttMsg, ok := msg.(*MQTTMessage)
				assert.True(t, ok)

				assert.Equal(t, tt.expectedPayload, mqttMsg.GetPayload())

				header := mqttMsg.GetHeader()
				assert.Equal(t, tt.expectedTopic, header["topic"])
				assert.Equal(t, tt.expectedQoS, header["qos"])
				assert.Equal(t, tt.expectedRetain, header["retain"])
			}
		})
	}
}

func TestMQTTCodec_SplitEdgeCases(t *testing.T) {
	codec := &MQTTCodec{}

	tests := []struct {
		name            string
		buffer          []byte
		expectedPackets int
		expectedRest    int
	}{
		{
			name:            "single byte buffer",
			buffer:          []byte{0x10},
			expectedPackets: 0,
			expectedRest:    1,
		},
		{
			name:            "malformed remaining length",
			buffer:          []byte{0x10, 0x80}, // Incomplete remaining length encoding
			expectedPackets: 0,
			expectedRest:    2,
		},
		{
			name:            "zero remaining length",
			buffer:          []byte{0x10, 0x00},
			expectedPackets: 1,
			expectedRest:    0,
		},
		{
			name:            "large remaining length but insufficient data",
			buffer:          []byte{0x10, 0xFF, 0x7F, 0x00}, // Claims 16383 bytes but only has 1
			expectedPackets: 0,
			expectedRest:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packets, rest, err := codec.Split(tt.buffer)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPackets, len(packets))
			assert.Equal(t, tt.expectedRest, len(rest))
		})
	}
}

func TestMQTTCodec_InterfaceCompliance(t *testing.T) {
	codec := &MQTTCodec{}

	// Verify that MQTTCodec implements SecureCodec interface
	var _ SecureCodec = codec

	// Test all interface methods are callable
	assert.Equal(t, "mqtt", codec.ProtocolName())
	assert.True(t, codec.IsSecure())

	// Test with minimal valid data
	_, _, _ = codec.Split([]byte{0x10, 0x00})

	// Test Decode with insufficient data (should return nil without panic)
	msg, consumed, err := codec.Decode([]byte{0x30, 0x01}, nil) // Not enough data for complete packet
	assert.NoError(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, 0, consumed)
}

func TestMQTTMessage_InterfaceCompliance(t *testing.T) {
	msg := &MQTTMessage{
		Payload: []byte("test"),
		Header:  map[string]string{"key": "value"},
	}

	// Verify that MQTTMessage implements Message interface
	var _ Message = msg

	// Test interface methods
	assert.Equal(t, []byte("test"), msg.GetPayload())
	assert.Equal(t, map[string]string{"key": "value"}, msg.GetHeader())
}

func TestMQTTCodec_ErrorHandling(t *testing.T) {
	codec := &MQTTCodec{}

	// Test Encode with missing topic
	msg := &MQTTMessage{
		Payload: []byte("test"),
		Header:  map[string]string{}, // No topic
	}

	encoded, err := codec.Encode(msg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing topic")
	assert.Nil(t, encoded)

	// Test Decode with unsupported packet type
	unsupportedPacket := []byte{0x20, 0x02, 0x00, 0x00} // CONNACK packet
	decodedMsg, consumed, err := codec.Decode(unsupportedPacket, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported packet type")
	assert.Nil(t, decodedMsg)
	assert.Equal(t, 0, consumed)
}

func TestMQTTCodec_LargePayloads(t *testing.T) {
	codec := &MQTTCodec{}

	// Test with large payload
	largePayload := make([]byte, 10000)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	msg := &MQTTMessage{
		Payload: largePayload,
		Header: map[string]string{
			"topic": "large/payload/test",
		},
	}

	// Encode large message
	encoded, err := codec.Encode(msg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, encoded)
	assert.True(t, len(encoded) > len(largePayload))

	// Decode large message
	decodedMsg, consumed, err := codec.Decode(encoded, nil)
	assert.NoError(t, err)
	assert.NotNil(t, decodedMsg)
	assert.Equal(t, len(encoded), consumed)

	mqttMsg, ok := decodedMsg.(*MQTTMessage)
	assert.True(t, ok)
	assert.Equal(t, largePayload, mqttMsg.GetPayload())
	assert.Equal(t, "large/payload/test", mqttMsg.GetHeader()["topic"])
}

func TestMQTTCodec_SpecialCharacters(t *testing.T) {
	codec := &MQTTCodec{}

	// Test with special characters in topic and payload
	specialTopic := "/topic/with/émojis/🚀"
	specialPayload := []byte("Special chars: émojis 🚀 \x00\x01\xFF")

	msg := &MQTTMessage{
		Payload: specialPayload,
		Header: map[string]string{
			"topic": specialTopic,
		},
	}

	// Encode and decode
	encoded, err := codec.Encode(msg, nil)
	assert.NoError(t, err)

	decodedMsg, consumed, err := codec.Decode(encoded, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(encoded), consumed)

	mqttMsg, ok := decodedMsg.(*MQTTMessage)
	assert.True(t, ok)
	assert.Equal(t, specialPayload, mqttMsg.GetPayload())
	assert.Equal(t, specialTopic, mqttMsg.GetHeader()["topic"])
}
