// Package codec provides a unified codec abstraction over different protocols. MQTTCodec implements SecureCodec interface.
package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// mqttConnect MQTT packet types.
	mqttConnect = 0x01
	// mqttPublish MQTT packet types.
	mqttPublish = 0x03
	// MQTT header shift.
	mqttHeaderShift = 4
)

// MQTTCodec implements SecureCodec interface.
type MQTTCodec struct{}

// Split splits the buffer into packets.
func (m *MQTTCodec) Split(buffer []byte) ([][]byte, []byte, error) {
	if len(buffer) < 2 {
		return nil, buffer, nil
	}

	remainingLength, n := DecodeRemainingLength(buffer[1:])
	if n == 0 || len(buffer) < 1+n {
		return nil, buffer, nil
	}

	totalLen := 1 + n + remainingLength
	if len(buffer) < totalLen {
		return nil, buffer, nil
	}

	packet := buffer[:totalLen]
	rest := buffer[totalLen:]
	return [][]byte{packet}, rest, nil
}

// Handshake parses the data and returns a Message, key, deviceID, and error.
func (m *MQTTCodec) Handshake(data []byte) (Message, []byte, string, error) {
	packets, _, err := m.Split(data)
	if err != nil || len(packets) == 0 {
		return nil, nil, "", errors.New("handshake packet invalid or incomplete")
	}
	packet := packets[0]

	packetType := packet[0] >> mqttHeaderShift
	if packetType != mqttConnect {
		return nil, nil, "", fmt.Errorf("expected CONNECT packet, got %d", packetType)
	}

	// Skip fixed header
	_, headerLen := DecodeRemainingLength(packet[1:])
	offset := 1 + headerLen

	// Protocol Name
	if len(packet) < offset+2 {
		return nil, nil, "", errors.New("invalid packet: missing protocol name length")
	}
	protoNameLen := int(binary.BigEndian.Uint16(packet[offset:]))
	offset += 2

	if len(packet) < offset+protoNameLen+4 {
		return nil, nil, "", errors.New("invalid protocol name section")
	}
	protoName := string(packet[offset : offset+protoNameLen])
	offset += protoNameLen

	protoLevel := packet[offset]
	offset++

	connectFlags := packet[offset]
	offset++

	keepAlive := binary.BigEndian.Uint16(packet[offset:])
	offset += 2

	// Payload - Client ID
	if len(packet) < offset+2 {
		return nil, nil, "", errors.New("invalid clientID length")
	}
	clientIDLen := int(binary.BigEndian.Uint16(packet[offset:]))
	offset += 2

	if len(packet) < offset+clientIDLen {
		return nil, nil, "", errors.New("clientID too short")
	}
	clientID := string(packet[offset : offset+clientIDLen])

	return &MQTTMessage{
		Payload: packet,
		Header: map[string]string{
			"protocol":  protoName,
			"version":   fmt.Sprintf("%d", protoLevel),
			"clientID":  clientID,
			"keepAlive": fmt.Sprintf("%d", keepAlive),
			"clean":     fmt.Sprintf("%t", connectFlags&0x02 > 0),
		},
	}, nil, clientID, nil
}

// Encode encodes the message into a byte array.
func (m *MQTTCodec) Encode(msg Message, _ []byte) ([]byte, error) {
	payload := msg.GetPayload()
	topic, ok := msg.GetHeader()["topic"]
	if !ok {
		return nil, errors.New("missing topic")
	}
	qos := byte(0)
	if q, ok := msg.GetHeader()["qos"]; ok && q == "1" {
		qos = 1
	}
	retain := byte(0)
	if r, ok := msg.GetHeader()["retain"]; ok && r == "1" {
		retain = 1
	}

	var fixedHeader byte = mqttPublish << mqttHeaderShift
	if retain > 0 {
		fixedHeader |= 0x01
	}
	if qos == 1 {
		fixedHeader |= 0x02
	}

	var body bytes.Buffer
	// Write topic
	length := len(topic)
	if length > 0xFFFF {
		return nil, fmt.Errorf("topic length exceeds uint16 max: %d", length)
	}
	_ = binary.Write(&body, binary.BigEndian, uint16(length))

	body.Write([]byte(topic))

	// Write payload
	body.Write(payload)

	// Build final packet
	var packet bytes.Buffer
	packet.WriteByte(fixedHeader)
	packet.Write(EncodeRemainingLength(body.Len()))
	packet.Write(body.Bytes())

	return packet.Bytes(), nil
}

// Decode decodes the data into a message.
func (m *MQTTCodec) Decode(data []byte, _ []byte) (Message, int, error) {
	if len(data) < 2 {
		return nil, 0, nil
	}

	packetType := data[0] >> mqttHeaderShift
	if packetType != mqttPublish {
		return nil, 0, errors.New("unsupported packet type")
	}

	remainingLength, n := DecodeRemainingLength(data[1:])
	totalLen := 1 + n + remainingLength
	if len(data) < totalLen {
		return nil, 0, nil
	}

	offset := 1 + n
	// topic name
	topicLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2
	topic := string(data[offset : offset+topicLen])
	offset += topicLen

	payload := data[offset:totalLen]

	retain := fmt.Sprintf("%t", data[0]&0x01 > 0)
	qos := fmt.Sprintf("%d", (data[0]>>1)&0x03)

	msg := &MQTTMessage{
		Payload: payload,
		Header: map[string]string{
			"topic":  topic,
			"retain": retain,
			"qos":    qos,
		},
	}
	return msg, totalLen, nil
}

// IsSecure returns true if the codec is secure.
func (m *MQTTCodec) IsSecure() bool {
	return true
}

// ProtocolName returns the name of the protocol.
func (m *MQTTCodec) ProtocolName() string {
	return "mqtt"
}

// MQTTMessage implements Message interface.
type MQTTMessage struct {
	Header  map[string]string
	Payload []byte
}

// GetPayload returns the payload of the message.
func (m *MQTTMessage) GetPayload() []byte {
	return m.Payload
}

// GetHeader returns the header of the message.
func (m *MQTTMessage) GetHeader() map[string]string {
	return m.Header
}

// EncodeRemainingLength encodes the remaining length into a byte array.
func EncodeRemainingLength(length int) []byte {
	var buf []byte
	for {
		d := byte(length % 128)
		length /= 128
		if length > 0 {
			d |= 0x80
		}
		buf = append(buf, d)
		if length == 0 {
			break
		}
	}
	return buf
}

// DecodeRemainingLength decodes the remaining length from a byte array.
func DecodeRemainingLength(buf []byte) (int, int) {
	var (
		value      = 0
		multiplier = 1
		i          = 0
	)
	for ; i < len(buf); i++ {
		d := int(buf[i])
		value += (d & 127) * multiplier
		if d&128 == 0 {
			return value, i + 1
		}
		multiplier *= 128
	}
	return 0, 0
}

// BuildMQTTConnectPacket builds a MQTT CONNECT packet.
func BuildMQTTConnectPacket(clientID string) []byte {
	packet := []byte{0x10}
	var header bytes.Buffer

	header.Write([]byte{
		0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol Name
		0x04,       // Protocol Level
		0x02,       // Clean session
		0x00, 0x3C, // KeepAlive = 60
	})

	clientIDBytes := []byte(clientID)
	header.Write([]byte{
		byte(len(clientIDBytes) >> 8), byte(len(clientIDBytes)),
	})
	header.Write(clientIDBytes)

	remaining := EncodeRemainingLength(header.Len())
	packet = append(packet, remaining...)
	packet = append(packet, header.Bytes()...)
	return packet
}
