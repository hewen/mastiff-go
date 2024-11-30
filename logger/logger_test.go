package logger

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestSetLevelError(t *testing.T) {
	err := InitLogger(Config{
		Level: "error level",
	})

	assert.NotNil(t, err)
}

func TestLogger(t *testing.T) {
	err := SetLevel(LogLevelInfo)
	assert.Nil(t, err)
	trace := NewTraceID()
	ctx := context.Background()
	ctx = context.WithValue(ctx, LoggerTraceKey, trace)

	testCase := []struct {
		l        *Logger
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
		assert.Equal(t, testCase[i].traceRes, l.GetTraceID() == trace)
		l.Debugf("tmp")
		l.Infof("tmp")
		l.Errorf("tmp")
		l.Warnf("tmp")
		l.Panicf("tmp")
		l.Fatalf("tmp")
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

	tmpFile, _ := os.CreateTemp("/tmp", "tmp.log")
	defer os.Remove(tmpFile.Name())

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
	ctx.Set(LoggerTraceKey, traceID)

	res = GetTraceIDWithGinContext(ctx)
	assert.Equal(t, traceID, res)
}

func TestNewLoggerWithGinContext(t *testing.T) {
	traceID := NewTraceID()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(LoggerTraceKey, traceID)

	l := NewLoggerWithGinContext(ctx)
	assert.Equal(t, traceID, l.traceID)
}

func TestNewOutgoingContextFromGinContext(t *testing.T) {
	traceID := NewTraceID()
	gctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	gctx.Set(LoggerTraceKey, traceID)

	ctx := NewOutgoingContextWithGinContext(gctx)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.Equal(t, true, ok)

	trace, ok := md[LoggerTraceKey]
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
	assert.Equal(t, traceID, l.traceID)

	ctx = NewOutgoingContextWithIncomingContext(context.TODO())
	l = NewLoggerWithContext(ctx)
	assert.NotEqual(t, traceID, l.traceID)
}
