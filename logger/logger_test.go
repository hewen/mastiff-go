package logger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestSetLevelError(t *testing.T) {
	err := InitLogger(Config{
		Level: "error level",
	})

	assert.NotNil(t, err)
}

func TestLogger(t *testing.T) {
	backends := []string{
		"std", "zap", "zerolog",
	}

	for i := range backends {
		err := InitLogger(Config{
			Backend: backends[i],
		})
		assert.Nil(t, err)
		trace := NewTraceID()
		ctx := context.Background()
		ctx = context.WithValue(ctx, LoggerTraceKey, trace)

		testCase := []struct {
			l        Logger
			traceRes bool
		}{
			{
				l:        NewLogger(),
				traceRes: false,
			},
			{
				l:        NewLoggerWithContext(context.Background()),
				traceRes: false,
			},
			{
				l:        NewLoggerWithTraceID(trace),
				traceRes: true,
			},
			{
				l:        NewLoggerWithContext(ctx),
				traceRes: true,
			},
		}
		for i := range testCase {
			l := testCase[i].l
			assert.Equal(t, testCase[i].traceRes, l.GetTraceID() == trace, fmt.Sprintf("case: %v logger trace: %v trace:%v", i, l.GetTraceID(), trace))
			l.Debugf("tmp")
			l.Infof("tmp")
			l.Errorf("tmp")
			l.Warnf("tmp")
			l.Panicf("tmp")
			l.Fatalf("tmp")
		}
	}
}

func TestInitLogger(t *testing.T) {
	err := InitLogger(Config{})
	assert.Nil(t, err)

	err = InitLogger(Config{
		Level:   "INFO",
		MaxSize: 100,
	})
	assert.Nil(t, err)

	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)

	err = InitLogger(Config{
		Level:   "INFO",
		Output:  tmpFile.Name(),
		MaxSize: 100,
	})
	assert.Nil(t, err)
}

func TestGetTraceIDWithGinContext(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	res := GetTraceIDWithGinContext(ctx)
	assert.Equal(t, true, res != "")

	traceID := NewTraceID()
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(string(LoggerTraceKey), traceID)

	res = GetTraceIDWithGinContext(ctx)
	assert.Equal(t, traceID, res)
}

func TestNewLoggerWithGinContext(t *testing.T) {
	traceID := NewTraceID()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(string(LoggerTraceKey), traceID)

	l := NewLoggerWithGinContext(ctx)
	assert.Equal(t, traceID, l.GetTraceID())
}

func TestNewOutgoingContextFromGinContext(t *testing.T) {
	traceID := NewTraceID()
	gctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gctx.Set(string(LoggerTraceKey), traceID)

	ctx := NewOutgoingContextWithGinContext(gctx)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.Equal(t, true, ok)

	trace, ok := md[string(LoggerTraceKey)]
	assert.Equal(t, true, ok)
	assert.Equal(t, traceID, trace[0])
}

func TestNewOutgoingContextFromIncomingContext(t *testing.T) {
	traceID := NewTraceID()
	m := make(map[string]string)
	m[string(LoggerTraceKey)] = traceID
	md := metadata.New(m)
	ictx := metadata.NewIncomingContext(context.TODO(), md)

	ctx := NewOutgoingContextWithIncomingContext(ictx)
	l := NewLoggerWithContext(ctx)
	assert.Equal(t, traceID, l.GetTraceID())

	ctx = NewOutgoingContextWithIncomingContext(context.TODO())
	l = NewLoggerWithContext(ctx)
	assert.NotEqual(t, traceID, l.GetTraceID())
}

func TestRotateAndLog(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)

	logger := &lumberjack.Logger{
		Filename: tmpFile.Name(),
	}

	rotateAndLog(logger)
}

type mockErrorLogger struct{}

func (m *mockErrorLogger) Rotate() error {
	return errors.New("mock rotate error")
}

func TestRotateAndLog_Error(t *testing.T) {
	logger := &mockErrorLogger{}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	rotateAndLog(logger)

	assert.Contains(t, buf.String(), "log rotation failed")
}

func TestConcurrentLogging(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "concurrent.log")
	_ = os.Remove(tmpFile)

	logger := &lumberjack.Logger{
		Filename:   tmpFile,
		MaxSize:    5, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   false,
	}

	defer func() {
		_ = logger.Close()
		_ = os.Remove(tmpFile)
	}()

	log.SetOutput(io.MultiWriter(logger, os.Stdout))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	const goroutines = 50
	const logsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				log.Printf("goroutine-%02d log line number %05d", id, j)
			}
		}(i)
	}

	wg.Wait()

	file, err := os.Open(tmpFile) // #nosec
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	linesCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "goroutine-") {
			t.Errorf("log line format unexpected: %s", line)
		}
		linesCount++
	}

	expectedLines := goroutines * logsPerGoroutine
	if linesCount < expectedLines {
		t.Errorf("log lines lost: got %d, expected at least %d", linesCount, expectedLines)
	} else {
		t.Logf("all logs written successfully: %d lines", linesCount)
	}
}
