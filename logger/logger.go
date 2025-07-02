// Package logger provides a simple logging interface with configurable levels and outputs.
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

// LogLevelFlag represents the log level as an integer flag.
type LogLevelFlag int

// LogLevel represents the log level as a string.
type LogLevel string

// contextKey is a type alias for string.
type contextKey string

const (
	// LoggerTraceKey is the key used to store the trace ID in the context.
	LoggerTraceKey contextKey = "logid"

	// LogLevelFlagFatal is the log level flag for fatal errors.
	LogLevelFlagFatal LogLevelFlag = 1
	// LogLevelFlagPanic is the log level flag for panic errors.
	LogLevelFlagPanic LogLevelFlag = 2
	// LogLevelFlagError is the log level flag for error messages.
	LogLevelFlagError LogLevelFlag = 3
	// LogLevelFlagWarn is the log level flag for warning messages.
	LogLevelFlagWarn LogLevelFlag = 4
	// LogLevelFlagInfo is the log level flag for informational messages.
	LogLevelFlagInfo LogLevelFlag = 5
	// LogLevelFlagDebug is the log level flag for debug messages.
	LogLevelFlagDebug LogLevelFlag = 6

	// LogLevelFatal is the log level for fatal errors.
	LogLevelFatal LogLevel = "FATAL"
	// LogLevelPanic is the log level for panic errors.
	LogLevelPanic LogLevel = "PANIC"
	// LogLevelError is the log level for error messages.
	LogLevelError LogLevel = "ERROR"
	// LogLevelWarn is the log level for warning messages.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelInfo is the log level for informational messages.
	LogLevelInfo LogLevel = "INFO"
	// LogLevelDebug is the log level for debug messages.
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

// InitLogger initializes the logger with the given configuration.
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

// SetLevel sets the log level for the logger.
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

// Logger is a simple logger that supports different log levels and trace IDs.
type Logger struct {
	traceID string
}

// GetTraceID returns the trace ID associated with the logger.
func (l *Logger) GetTraceID() string {
	return l.traceID
}

// Debugf logs a debug message with the given format and arguments.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.output(LogLevelFlagDebug, format, v...)
}

// Infof logs an info message with the given format and arguments.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(LogLevelFlagInfo, format, v...)
}

// Warnf logs a warning message with the given format and arguments.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.output(LogLevelFlagWarn, format, v...)
}

// Errorf logs an error message with the given format and arguments.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(LogLevelFlagError, format, v...)
}

// Panicf logs a panic message with the given format and arguments, and then panics.
func (l *Logger) Panicf(format string, v ...interface{}) {
	l.output(LogLevelFlagPanic, format, v...)
}

// Fatalf logs a fatal message with the given format and arguments, and then exits the program.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(LogLevelFlagFatal, format, v...)
}

// output writes a log message with the specified level and format to the logger.
func (l *Logger) output(level LogLevelFlag, format string, v ...interface{}) {
	if level > logLevel {
		return
	}

	log.Printf("["+l.traceID+"] | "+string(logLevelValueMap[level])+" | "+
		format, v...)
}

// NewTraceID generates a new trace ID using the shortid package.
func NewTraceID() string {
	traceID, _ := shortid.Generate()
	return traceID
}

// NewLogger creates a new Logger instance with a generated trace ID.
func NewLogger() *Logger {
	return &Logger{
		traceID: NewTraceID(),
	}
}

// NewLoggerWithTraceID creates a new Logger instance with the specified trace ID.
func NewLoggerWithTraceID(traceID string) *Logger {
	return &Logger{
		traceID: traceID,
	}
}

// NewLoggerWithContext creates a new Logger instance using the trace ID from the provided context.
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

// GetTraceIDWithGinContext retrieves the trace ID from the Gin context.
func GetTraceIDWithGinContext(ctx *gin.Context) string {
	res, exists := ctx.Get(string(LoggerTraceKey))
	if exists {
		return res.(string)
	}

	return NewTraceID()
}

// NewLoggerWithGinContext creates a new Logger instance using the trace ID from the Gin context.
func NewLoggerWithGinContext(ctx *gin.Context) *Logger {
	return NewLoggerWithTraceID(GetTraceIDWithGinContext(ctx))
}

// NewOutgoingContextWithGinContext creates a new outgoing context with the trace ID from the Gin context.
func NewOutgoingContextWithGinContext(ctx *gin.Context) context.Context {
	m := make(map[string]string)
	m[string(LoggerTraceKey)] = GetTraceIDWithGinContext(ctx)
	md := metadata.New(m)
	return metadata.NewOutgoingContext(context.TODO(), md)
}

// NewOutgoingContextWithIncomingContext creates a new outgoing context with the trace ID from the incoming context.
func NewOutgoingContextWithIncomingContext(ctx context.Context) context.Context {
	var traceID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if t, ok := md[string(LoggerTraceKey)]; ok && len(t) > 0 {
			traceID = t[0]
		}
	}

	if traceID == "" {
		traceID = NewTraceID()
	}

	m := make(map[string]string)
	md := metadata.New(m)
	m[string(LoggerTraceKey)] = traceID
	ctx = metadata.NewOutgoingContext(ctx, md)

	return context.WithValue(ctx, LoggerTraceKey, traceID)
}
