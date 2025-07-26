// Package codec provides a unified codec abstraction over different protocols.
package codec

// SecureCodec represents a secure codec that can encode/decode messages.
type SecureCodec interface {
	Split(buffer []byte) ([][]byte, []byte, error)
	Handshake(data []byte) (msg Message, key []byte, deviceID string, err error)

	Encode(msg Message, key []byte) ([]byte, error)
	Decode(data []byte, key []byte) (msg Message, consumed int, err error)

	IsSecure() bool
	ProtocolName() string
}

// Message represents a protocol message.
type Message interface {
	GetPayload() []byte
	GetHeader() map[string]string
}
