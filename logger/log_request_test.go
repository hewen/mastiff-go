package logger

import (
	"errors"
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

func (m *mockLogger) Infof(format string, _ ...any) {
	m.lastLevel = "info"
	m.lastMsg = format
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
}
