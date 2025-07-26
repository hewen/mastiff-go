package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConn is a mock implementation of the Conn interface.
type MockConn struct {
	mock.Mock
}

func (m *MockConn) AsyncWrite(data []byte, callback AsyncCallback) error {
	args := m.Called(data, callback)
	return args.Error(0)
}

func (m *MockConn) Write(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) Context() any {
	args := m.Called()
	return args.Get(0)
}

func (m *MockConn) SetContext(ctx any) {
	m.Called(ctx)
}

// nolint
func TestSetContextValue(t *testing.T) {
	tests := []struct {
		expectedCalls  int
		initialContext any
		value          any
		name           string
		key            string
	}{
		{
			name:           "set value on empty context",
			initialContext: nil,
			key:            "test_key",
			value:          "test_value",
			expectedCalls:  1,
		},
		{
			name:           "set value on existing context",
			initialContext: map[string]any{"existing": "value"},
			key:            "new_key",
			value:          "new_value",
			expectedCalls:  1,
		},
		{
			name:           "overwrite existing value",
			initialContext: map[string]any{"test_key": "old_value"},
			key:            "test_key",
			value:          "new_value",
			expectedCalls:  1,
		},
		{
			name:           "set value on non-map context",
			initialContext: "not a map",
			key:            "test_key",
			value:          "test_value",
			expectedCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockConn)
			mockConn.On("Context").Return(tt.initialContext)
			mockConn.On("SetContext", mock.AnythingOfType("map[string]interface {}")).Return()

			SetContextValue(mockConn, tt.key, tt.value)

			mockConn.AssertNumberOfCalls(t, "Context", 1)
			mockConn.AssertNumberOfCalls(t, "SetContext", tt.expectedCalls)

			// Verify the context was set correctly
			calls := mockConn.Calls
			if len(calls) >= 2 {
				setContextCall := calls[1]
				ctx := setContextCall.Arguments[0].(map[string]any)
				assert.Equal(t, tt.value, ctx[tt.key])
			}
		})
	}
}

func TestGetContextValue_IntType(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Context").Return(map[string]any{"int_key": 42})

	value, exists := GetContextValue[int](mockConn, "int_key")
	assert.Equal(t, 42, value)
	assert.True(t, exists)
}

func TestGetContextValue_BoolType(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Context").Return(map[string]any{"bool_key": true})

	value, exists := GetContextValue[bool](mockConn, "bool_key")
	assert.Equal(t, true, value)
	assert.True(t, exists)
}

func TestContextKeyConstants(t *testing.T) {
	assert.Equal(t, contextKey("device_id"), ContextKeyDeviceID)
}

func TestGetContextValue_ErrorType(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Context").Return(map[string]any{"int_key": 42})

	_, exists := GetContextValue[string](mockConn, "int_key")
	assert.False(t, exists)
}

func TestGetContextValue_NotExist(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Context").Return(map[string]any{"int_key": 42})

	_, exists := GetContextValue[int](mockConn, "key")
	assert.False(t, exists)
}
