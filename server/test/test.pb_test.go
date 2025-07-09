package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

func TestTestMsg_FieldAccess(t *testing.T) {
	msg := &TestMsg{
		Id:   123,
		Name: "Alice",
	}

	assert.EqualValues(t, int32(123), msg.GetId(), "ID should be 123")
	assert.EqualValues(t, "Alice", msg.GetName(), "Name should be Alice")
}

func TestTestMsg_Reset(t *testing.T) {
	msg := &TestMsg{
		Id:   123,
		Name: "Alice",
	}
	msg.Reset()

	assert.EqualValues(t, int32(0), msg.GetId(), "ID should be reset to 0")
	assert.EqualValues(t, "", msg.GetName(), "Name should be reset to empty string")
}

func TestTestMsg_Descriptor(t *testing.T) {
	msg := &TestMsg{
		Id:   123,
		Name: "Alice",
	}
	_, _ = msg.Descriptor()

	desc := msg.ProtoReflect().Descriptor()

	assert.Equal(t, protoreflect.Name("TestMsg"), desc.Name())
	assert.Equal(t, protoreflect.FullName("test.TestMsg"), desc.FullName())
	assert.Equal(t, 2, desc.Fields().Len(), "Should have exactly 2 fields")

	idField := desc.Fields().ByName("id")
	assert.NotNil(t, idField)
	assert.Equal(t, protoreflect.Int32Kind, idField.Kind())

	nameField := desc.Fields().ByName("name")
	assert.NotNil(t, nameField)
	assert.Equal(t, protoreflect.StringKind, nameField.Kind())
}

func TestTestMsg_String(t *testing.T) {
	msg := &TestMsg{Id: 1, Name: "Bob"}
	str := msg.String()

	assert.NotEmpty(t, str, "String representation should not be empty")
}

func TestTestMsg_MarshalUnmarshal(t *testing.T) {
	original := &TestMsg{
		Id:   42,
		Name: "Charlie",
	}

	data, err := proto.Marshal(original)
	assert.NoError(t, err, "Marshalling should not return error")
	assert.NotNil(t, data, "Serialized data should not be nil")

	var decoded TestMsg
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err, "Unmarshalling should not return error")
	assert.True(t, proto.Equal(original, &decoded), "Decoded message should match original")
}
