package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/teris-io/shortid"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevelFlag int
type LogLevel string

const (
	LoggerTraceKey = "logid"

	LogLevelFlagFatal LogLevelFlag = 1
	LogLevelFlagPanic LogLevelFlag = 2
	LogLevelFlagError LogLevelFlag = 3
	LogLevelFlagWarn  LogLevelFlag = 4
	LogLevelFlagInfo  LogLevelFlag = 5
	LogLevelFlagDebug LogLevelFlag = 6

	LogLevelFatal LogLevel = "FATAL"
	LogLevelPanic LogLevel = "PANIC"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
)

var (
	logLevel = LogLevelFlagInfo

	logLevelMap = map[LogLevel]LogLevelFlag{
		LogLevelFatal: LogLevelFlagFatal,
		LogLevelPanic: LogLevelFlagPanic,
		LogLevelError: LogLevelFlagError,
		LogLevelWarn:  LogLevelFlagWarn,
		LogLevelInfo:  LogLevelFlagInfo,
		LogLevelDebug: LogLevelFlagDebug,
	}

	logLevelValueMap = map[LogLevelFlag]LogLevel{
		LogLevelFlagFatal: LogLevelFatal,
		LogLevelFlagPanic: LogLevelPanic,
		LogLevelFlagError: LogLevelError,
		LogLevelFlagWarn:  LogLevelWarn,
		LogLevelFlagInfo:  LogLevelInfo,
		LogLevelFlagDebug: LogLevelDebug,
	}
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func InitLogger(conf Config) error {
	if err := SetLevel(conf.Level); err != nil {
		return err
	}

	if conf.Output == "" {
		return nil
	}

	log.SetOutput(io.MultiWriter(&lumberjack.Logger{
		Filename: conf.Output,
		MaxSize:  conf.MaxSize,
	}, os.Stdout))

	return nil
}

func SetLevel(level LogLevel) error {
	logLevel = LogLevelFlagInfo
	if level == "" {
		return nil
	}

	var ok bool
	logLevel, ok = logLevelMap[level]
	if !ok {
		return fmt.Errorf("log level set error")
	}
	return nil
}

type Logger struct {
	traceID string
}

func (l *Logger) GetTraceID() string {
	return l.traceID
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.output(LogLevelFlagDebug, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(LogLevelFlagInfo, format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.output(LogLevelFlagWarn, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(LogLevelFlagError, format, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.output(LogLevelFlagPanic, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(LogLevelFlagFatal, format, v...)
}

func (l *Logger) output(level LogLevelFlag, format string, v ...interface{}) {
	if level > logLevel {
		return
	}

	log.Printf("["+l.traceID+"] | "+string(logLevelValueMap[level])+" | "+
		format, v...)
}

func NewTraceID() string {
	traceID, _ := shortid.Generate()
	return traceID
}

func NewLogger() *Logger {
	return &Logger{
		traceID: NewTraceID(),
	}
}

func NewLoggerWithTraceID(traceID string) *Logger {
	return &Logger{
		traceID: traceID,
	}
}

func NewLoggerWithContext(ctx context.Context) *Logger {
	val := ctx.Value(LoggerTraceKey)
	if traceID, ok := val.(string); ok {
		return &Logger{
			traceID: traceID,
		}

	}

	return &Logger{
		traceID: NewTraceID(),
	}
}

func GetTraceIDWithGinContext(ctx *gin.Context) string {
	res, exists := ctx.Get(LoggerTraceKey)
	if exists {
		return res.(string)
	}

	return NewTraceID()
}

func NewLoggerWithGinContext(ctx *gin.Context) *Logger {
	return NewLoggerWithTraceID(GetTraceIDWithGinContext(ctx))
}

func NewOutgoingContextWithGinContext(ctx *gin.Context) context.Context {
	m := make(map[string]string)
	m[string(LoggerTraceKey)] = GetTraceIDWithGinContext(ctx)
	md := metadata.New(m)
	return metadata.NewOutgoingContext(context.TODO(), md)
}

func NewOutgoingContextWithIncomingContext(ctx context.Context) context.Context {
	var traceID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if t, ok := md[LoggerTraceKey]; ok && len(t) > 0 {
			traceID = t[0]
		}
	}

	if traceID == "" {
		traceID = NewTraceID()
	}

	m := make(map[string]string)
	md := metadata.New(m)
	m[LoggerTraceKey] = traceID
	ctx = metadata.NewOutgoingContext(ctx, md)

	return context.WithValue(ctx, LoggerTraceKey, traceID)
}
