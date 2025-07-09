package logger

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	traceID   string
	lastLevel string
	lastMsg   string
	fields    map[string]any
}

func (m *mockLogger) GetTraceID() string {
	return m.traceID
}

func (m *mockLogger) Debugf(format string, _ ...any) {
	m.lastLevel = "debug"
	m.lastMsg = format
}

func (m *mockLogger) Infof(format string, v ...any) {
	m.lastLevel = "info"
	m.lastMsg = format
	NewLogger().Infof(format, v...)
}

func (m *mockLogger) Warnf(format string, _ ...any) {
	m.lastLevel = "warn"
	m.lastMsg = format
}

func (m *mockLogger) Errorf(format string, _ ...any) {
	m.lastLevel = "error"
	m.lastMsg = format
}

func (m *mockLogger) Panicf(format string, _ ...any) {
	m.lastLevel = "panic"
	m.lastMsg = format
}

func (m *mockLogger) Fatalf(format string, _ ...any) {
	m.lastLevel = "fatal"
	m.lastMsg = format
}

func (m *mockLogger) Fields(f map[string]any) Logger {
	m.fields = f
	return m
}

func TestLogRequest(t *testing.T) {
	enableMasking = false

	t.Run("Normal request", func(t *testing.T) {
		mock := &mockLogger{}
		LogRequest(mock, 200, 500*time.Millisecond, "127.0.0.1", "GET /ping", "curl/1.0", "req-body", "resp-body", nil)

		assert.Equal(t, "info", mock.lastLevel)
		assert.Equal(t, "req", mock.lastMsg)
		assert.Equal(t, 200, mock.fields["status"])
		assert.Equal(t, "127.0.0.1", mock.fields["ip"])
		assert.Equal(t, "GET /ping", mock.fields["method"])
		assert.Equal(t, "curl/1.0", mock.fields["ua"])
		assert.Equal(t, "req-body", mock.fields["req"])
		assert.Equal(t, "resp-body", mock.fields["resp"])
	})

	t.Run("Slow request", func(t *testing.T) {
		mock := &mockLogger{}
		LogRequest(mock, 200, 2*time.Second, "127.0.0.1", "GET /slow", "test-agent", "req", "resp", nil)

		assert.Equal(t, "info", mock.lastLevel)
		assert.Equal(t, "slow req", mock.lastMsg)
	})

	t.Run("Request with error", func(t *testing.T) {
		mock := &mockLogger{}
		err := errors.New("db error")
		LogRequest(mock, 500, 100*time.Millisecond, "127.0.0.1", "POST /api", "agent", "req", "resp", err)

		assert.Equal(t, "error", mock.lastLevel)
		assert.Equal(t, "req", mock.lastMsg)
		assert.Equal(t, "db error", mock.fields["err"])
	})

	t.Run("Request with nil", func(t *testing.T) {
		mock := &mockLogger{}
		LogRequest(mock, 500, 100*time.Millisecond, "127.0.0.1", "POST /api", "agent", "nil", "nil", nil)

		assert.Equal(t, "req", mock.lastMsg)
	})
}

type TestPayload struct {
	Name     string
	Mobile   string `mask:"mobile"`
	Password string `mask:"password"`
	Email    string `mask:"email"`
}

type TestNoMaskPayload struct {
	Name     string
	Mobile   string
	Password string
	Email    string
}

var req = TestPayload{
	Name:     "Alice",
	Mobile:   "13912345678",
	Password: "SuperSecret",
	Email:    "alice@example.com",
}

var resp = TestPayload{
	Name:     "Alice",
	Mobile:   "13912345678",
	Password: "SuperSecret",
	Email:    "alice@example.com",
}

func BenchmarkLogRequestWithoutMask(b *testing.B) {
	SetLogMasking(false)
	l := NewLogger()
	fmt.Println("")
	for i := 0; i < b.N; i++ {
		LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", req, resp, nil)
	}
}

func BenchmarkLogRequestWithMask(b *testing.B) {
	SetLogMasking(true)
	l := NewLogger()
	fmt.Println("")
	for i := 0; i < b.N; i++ {
		LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", req, resp, nil)
	}
}

func TestLogRequestWithMask(_ *testing.T) {
	SetLogMasking(true)
	// test repeat log
	l := NewLogger()
	LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", req, resp, nil)
	LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", req, resp, nil)

	LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", nil, nil, nil)
	LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", []string{"test"}, nil, nil)
	LogRequest(l, 200, 300*time.Millisecond, "127.0.0.1", "POST /test", "Go-http-client/1.1", 1, nil, nil)
}

func TestMaskValue(t *testing.T) {
	SetLogMasking(true)
	res := MaskValue("test")
	assert.Equal(t, "test", res)

	res = MaskValue(nil)
	assert.Equal(t, nil, res)

	res = MaskValue(1)
	assert.Equal(t, 1, res)

	res = MaskValue(req)
	assert.EqualValues(t, &TestPayload{
		Name:     "Alice",
		Mobile:   "1391***5678",
		Password: "**************",
		Email:    "ali****@example.com",
	}, res)

}
