package logger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/hewen/mastiff-go/config/loggerconf"
	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestSetLevelError(t *testing.T) {
	err := InitLogger(loggerconf.Config{
		Level: "error level",
	})

	assert.NotNil(t, err)
}

func TestLogger(t *testing.T) {
	defer func() {
		_ = recover()
	}()

	backends := []string{
		"std", "zap", "zerolog",
	}

	for i := range backends {
		err := InitLogger(loggerconf.Config{
			Backend: backends[i],
			Level:   LogLevelDebug,
			Outputs: []string{"stdout", "stderr", "errorout"},
		})
		assert.Nil(t, err)
		trace := NewTraceID()
		ctx := context.Background()
		ctx = context.WithValue(ctx, contextkeys.LoggerTraceIDKey, trace)

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
		for j := range testCase {
			l := testCase[j].l
			assert.Equal(t, testCase[j].traceRes, l.GetTraceID() == trace, fmt.Sprintf("case: %v logger trace: %v trace:%v", i, l.GetTraceID(), trace))
			data := map[string]any{
				"backend": backends[i],
			}

			entry := l.Fields(data)
			entry.Debugf("tmp")
			entry.Infof("tmp")
			entry.Errorf("tmp")
			entry.Warnf("tmp")
		}
	}
}

func TestStdLoggerPanicAndFatalf(_ *testing.T) {
	_ = InitLogger(loggerconf.Config{
		Backend: "std",
	})
	NewLogger().Panicf("tmp")
	NewLogger().Fatalf("tmp")
}

func TestZapLoggerPanic(_ *testing.T) {
	defer func() {
		_ = recover()
	}()
	_ = InitLogger(loggerconf.Config{
		Backend: "zap",
	})
	NewLogger().Panicf("test")
}

func TestZerologLoggerPanic(_ *testing.T) {
	defer func() {
		_ = recover()
	}()
	_ = InitLogger(loggerconf.Config{
		Backend: "zerolog",
	})
	NewLogger().Panicf("test")
}

func TestZapLoggerFatalf(t *testing.T) {
	const fatalEnv = "TEST_FATAL"
	if os.Getenv(fatalEnv) == "1" {
		_ = InitLogger(loggerconf.Config{
			Backend: "zap",
			Level:   LogLevelDebug,
		})
		NewLogger().Fatalf("fatal test")
		return
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to get executable path: %v", err)
	}

	cmd := exec.Command(exe, "-test.run=TestZapLoggerFatalf") // #nosec
	cmd.Env = append(os.Environ(), fatalEnv+"=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		t.Fatal("expected Fatalf to call os.Exit(1), but it did not")
	}
}

func TestZerologLoggerFatalf(t *testing.T) {
	const fatalEnv = "TEST_FATAL"
	if os.Getenv(fatalEnv) == "1" {
		_ = InitLogger(loggerconf.Config{
			Backend: "zerolog",
			Level:   LogLevelDebug,
		})
		NewLogger().Fatalf("fatal test")
		return
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to get executable path: %v", err)
	}

	cmd := exec.Command(exe, "-test.run=TestZerologLoggerFatalf") // #nosec
	cmd.Env = append(os.Environ(), fatalEnv+"=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		t.Fatal("expected Fatalf to call os.Exit(1), but it did not")
	}
}

func TestInitLogger(t *testing.T) {
	err := InitLogger(loggerconf.Config{})
	assert.Nil(t, err)

	err = InitLogger(loggerconf.Config{
		Level: LogLevelInfo,
	})
	assert.Nil(t, err)

	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(loggerconf.Config{
		Level:   LogLevelInfo,
		Outputs: []string{"file"},
		FileOutput: &loggerconf.FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(t, err)
}

func TestNewOutgoingContextFromIncomingContext(t *testing.T) {
	traceID := NewTraceID()
	m := make(map[string]string)
	m[string(contextkeys.LoggerTraceIDKey)] = traceID
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

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

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
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)
	logger := &lumberjack.Logger{
		Filename:   tmpFile.Name(),
		MaxSize:    5, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   false,
	}

	defer func() {
		_ = logger.Close()
		_ = os.Remove(tmpFile.Name())
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

	scanner := bufio.NewScanner(tmpFile)
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

func TestValidate(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	tests := []struct {
		name    string
		conf    loggerconf.Config
		wantErr bool
	}{
		{
			name: "valid config with file",
			conf: loggerconf.Config{
				Outputs: []string{"file"},
				FileOutput: &loggerconf.FileOutputConfig{
					Path: tmpFile.Name(),
				},
				Backend: "zerolog",
			},
			wantErr: false,
		},
		{
			name: "file output missing path",
			conf: loggerconf.Config{
				Outputs:    []string{"file"},
				FileOutput: &loggerconf.FileOutputConfig{}, // missing Path
				Backend:    "zap",
			},
			wantErr: true,
		},
		{
			name: "invalid backend",
			conf: loggerconf.Config{
				Outputs: []string{"stdout"},
				Backend: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.Validate()
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}

func TestCreateFileWriter(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	tests := []struct {
		policy string
	}{
		{"daily"},
		{"size"},
		{"none"},
		{"invalid"}, // should fallback to daily
	}

	for _, tt := range tests {
		t.Run(tt.policy, func(t *testing.T) {
			cfg := loggerconf.FileOutputConfig{
				Path:         tmpFile.Name(),
				RotatePolicy: tt.policy,
				MaxSize:      1,
			}
			writer := createFileWriter(cfg)
			assert.NotNil(t, writer)
		})
	}
}

func TestNewSizeLogger(t *testing.T) {
	cfg := loggerconf.FileOutputConfig{
		Path:    "/tmp/test-size.log",
		MaxSize: 5,
	}
	l := newSizeLogger(cfg)
	assert.NotNil(t, l)
}

func TestNewPlainFileLogger_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	cfg := loggerconf.FileOutputConfig{Path: tmpFile.Name()}
	w := newPlainFileLogger(cfg)
	assert.NotNil(t, w)
}

func TestNewPlainFileLogger_Failure(t *testing.T) {
	cfg := loggerconf.FileOutputConfig{Path: os.TempDir() + "errordir/forbidden.log"}
	w := newPlainFileLogger(cfg)
	assert.NotNil(t, w)
}

func TestInitLogger_ErrorBackend(t *testing.T) {
	err := InitLogger(loggerconf.Config{
		Backend: "error",
	})

	assert.NotNil(t, err)
}

func TestLogerLever(t *testing.T) {
	err := SetLevel(LogLevelError)
	assert.Nil(t, err)
	l := NewLogger()
	l.Infof("not display")
	l.Errorf("display")
}
