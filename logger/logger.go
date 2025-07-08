// Package logger provides a simple logging interface with configurable levels and outputs.
// It supports multiple backend implementations (std, zap, zerolog), trace ID propagation,
// daily log rotation, and context integration for Gin and gRPC.
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevelFlag represents the log level as an integer flag.
type LogLevelFlag int

// LogLevel represents the log level as a string.
type LogLevel string

// contextKey is a type alias for string used as context key type.
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

	// LogLevelFatal is the log level string for fatal errors.
	LogLevelFatal LogLevel = "FATAL"
	// LogLevelPanic is the log level string for panic errors.
	LogLevelPanic LogLevel = "PANIC"
	// LogLevelError is the log level string for error messages.
	LogLevelError LogLevel = "ERROR"
	// LogLevelWarn is the log level string for warning messages.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelInfo is the log level string for informational messages.
	LogLevelInfo LogLevel = "INFO"
	// LogLevelDebug is the log level string for debug messages.
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

// Logger interface defines the common logging methods with traceID support.
type Logger interface {
	// GetTraceID returns the trace ID associated with the logger.
	GetTraceID() string

	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Panicf(format string, v ...any)
	Fatalf(format string, v ...any)
}

var (
	// defaultLogger is the default logger instance.
	defaultLogger Logger
)

// InitLogger initializes the global logger with the given configuration.
func InitLogger(conf Config) error {
	err := SetLevel(conf.Level)
	if err != nil {
		return err
	}

	var out io.Writer = os.Stdout
	if conf.Output != "" {
		rotator := newDailyRotatingLogger(conf)
		out = io.MultiWriter(out, rotator)
	}

	traceID := NewTraceID()
	switch conf.Backend {
	case "zap":
		var zapLog *zap.Logger
		zapLog = newZapLogger(out)
		defaultLogger = &zapLogger{logger: zapLog.Sugar(), traceID: traceID}
	case "zerolog":
		zl := newZeroLogger(out)
		defaultLogger = &zerologLogger{logger: zl, traceID: traceID}
	default:
		// std
		stdLog := newStdLogger(out)
		defaultLogger = &stdLogger{logger: stdLog, traceID: traceID}
	}
	return err
}

func newDailyRotatingLogger(conf Config) *lumberjack.Logger {
	logger := &lumberjack.Logger{
		Filename: conf.Output,
		MaxSize:  conf.MaxSize,
	}

	c := cron.New(cron.WithSeconds())
	_, _ = c.AddFunc("0 0 0 * * *", func() {
		rotateAndLog(logger)
	})
	c.Start()
	return logger
}

// rotatable interface defines the Rotate method for log rotation.
type rotatable interface {
	Rotate() error
}

// rotateAndLog rotates the log file and logs any errors.
func rotateAndLog(rotator rotatable) {
	if e := rotator.Rotate(); e != nil {
		log.Printf("log rotation failed: %v", e)
	}
}

// newZapLogger creates a new zap logger with the specified output.
func newZapLogger(out io.Writer) *zap.Logger {
	writeSyncer := zapcore.AddSync(out)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, writeSyncer, zap.InfoLevel)

	logger := zap.New(core, zap.AddCaller())

	return logger
}

// newZeroLogger creates a new zerolog logger with the specified output.
func newZeroLogger(out io.Writer) zerolog.Logger {
	return zerolog.New(out).With().Timestamp().Logger()
}

// newStdLogger creates a new standard logger with the specified output.
func newStdLogger(out io.Writer) *log.Logger {
	return log.New(out, "", log.Ldate|log.Lmicroseconds)
}

// SetLevel sets the global logging level.
func SetLevel(level LogLevel) error {
	if level == "" {
		logLevel = LogLevelFlagInfo
		return nil
	}
	lv, ok := logLevelMap[level]
	if !ok {
		return fmt.Errorf("invalid log level: %s", level)
	}
	logLevel = lv
	return nil
}

// NewTraceID generates a new trace ID using the shortid package.
func NewTraceID() string {
	id, _ := shortid.Generate()
	return id
}

// NewLogger creates a new Logger instance with a generated trace ID.
func NewLogger() Logger {
	return NewLoggerWithTraceID(NewTraceID())
}

// NewLoggerWithContext returns a new Logger with trace ID extracted from context.
func NewLoggerWithContext(ctx context.Context) Logger {
	return NewLoggerWithTraceID(GetTraceIDWithContext(ctx))
}

// GetTraceIDWithGinContext returns the trace ID from Gin context or generates a new one.
func GetTraceIDWithGinContext(ctx *gin.Context) string {
	if v, exists := ctx.Get(string(LoggerTraceKey)); exists {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return NewTraceID()
}

// GetTraceIDWithContext returns the trace ID from context or generates a new one.
func GetTraceIDWithContext(ctx context.Context) string {
	if v, exists := ctx.Value(LoggerTraceKey).(string); exists {
		return v
	}
	return NewTraceID()
}

// NewLoggerWithGinContext returns a new Logger with trace ID extracted from Gin context.
func NewLoggerWithGinContext(ctx *gin.Context) Logger {
	return NewLoggerWithTraceID(GetTraceIDWithGinContext(ctx))
}

// NewLoggerWithTraceID returns a new Logger with the specified trace ID.
func NewLoggerWithTraceID(traceID string) Logger {
	// Return a logger based on the global defaultLogger but override traceID.
	switch v := defaultLogger.(type) {
	case *stdLogger:
		return &stdLogger{logger: v.logger, traceID: traceID}
	case *zapLogger:
		return &zapLogger{logger: v.logger, traceID: traceID}
	case *zerologLogger:
		return &zerologLogger{logger: v.logger, traceID: traceID}
	default:
		return &stdLogger{logger: log.Default(), traceID: traceID}
	}
}

// NewOutgoingContextWithIncomingContext creates a new outgoing context containing trace ID from incoming context metadata.
func NewOutgoingContextWithIncomingContext(ctx context.Context) context.Context {
	var traceID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if tid, ok := md[string(LoggerTraceKey)]; ok && len(tid) > 0 {
			traceID = tid[0]
		}
	}
	if traceID == "" {
		traceID = NewTraceID()
	}
	md := metadata.Pairs(string(LoggerTraceKey), traceID)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return context.WithValue(ctx, LoggerTraceKey, traceID)
}

// NewOutgoingContextWithGinContext creates a new outgoing context with the trace ID from the Gin context.
func NewOutgoingContextWithGinContext(ctx *gin.Context) context.Context {
	m := make(map[string]string)
	m[string(LoggerTraceKey)] = GetTraceIDWithGinContext(ctx)
	md := metadata.New(m)
	return metadata.NewOutgoingContext(context.TODO(), md)

}

// stdLogger is a Logger implementation using the standard log package.
type stdLogger struct {
	logger  *log.Logger
	traceID string
}

func (l *stdLogger) GetTraceID() string { return l.traceID }

func (l *stdLogger) logOutput(level LogLevelFlag, format string, v ...any) {
	if level > logLevel {
		return
	}
	prefix := fmt.Sprintf("[%s] [%s]", l.GetTraceID(), logLevelValueMap[level])
	_ = l.logger.Output(3, prefix+" "+fmt.Sprintf(format, v...))
}

func (l *stdLogger) Debugf(format string, v ...any) { l.logOutput(LogLevelFlagDebug, format, v...) }
func (l *stdLogger) Infof(format string, v ...any)  { l.logOutput(LogLevelFlagInfo, format, v...) }
func (l *stdLogger) Warnf(format string, v ...any)  { l.logOutput(LogLevelFlagWarn, format, v...) }
func (l *stdLogger) Errorf(format string, v ...any) { l.logOutput(LogLevelFlagError, format, v...) }
func (l *stdLogger) Panicf(format string, v ...any) { l.logOutput(LogLevelFlagPanic, format, v...) }
func (l *stdLogger) Fatalf(format string, v ...any) { l.logOutput(LogLevelFlagFatal, format, v...) }

// zapLogger is a Logger implementation using Uber's zap package.
type zapLogger struct {
	logger  *zap.SugaredLogger
	traceID string
}

func (l *zapLogger) GetTraceID() string { return l.traceID }

func (l *zapLogger) logOutput(level LogLevelFlag, format string, v ...any) {
	if level > logLevel {
		return
	}
	l.logger.
		With("loglevel", string(logLevelValueMap[level])).
		With("trace", l.GetTraceID()).
		Infof(format, v...)
}

func (l *zapLogger) Debugf(format string, v ...any) { l.logOutput(LogLevelFlagDebug, format, v...) }
func (l *zapLogger) Infof(format string, v ...any)  { l.logOutput(LogLevelFlagInfo, format, v...) }
func (l *zapLogger) Warnf(format string, v ...any)  { l.logOutput(LogLevelFlagWarn, format, v...) }
func (l *zapLogger) Errorf(format string, v ...any) { l.logOutput(LogLevelFlagError, format, v...) }
func (l *zapLogger) Panicf(format string, v ...any) { l.logOutput(LogLevelFlagPanic, format, v...) }
func (l *zapLogger) Fatalf(format string, v ...any) { l.logOutput(LogLevelFlagFatal, format, v...) }

// zerologLogger is a Logger implementation using the zerolog package.
type zerologLogger struct {
	logger  zerolog.Logger
	traceID string
}

func (l *zerologLogger) GetTraceID() string { return l.traceID }

func (l *zerologLogger) logOutput(level LogLevelFlag, format string, v ...any) {
	if level > logLevel {
		return
	}
	l.logger.Info().
		Str("loglevel", string(logLevelValueMap[level])).
		Str("strace", l.GetTraceID()).
		Msgf(format, v...)
}

func (l *zerologLogger) Debugf(format string, v ...any) { l.logOutput(LogLevelFlagDebug, format, v...) }
func (l *zerologLogger) Infof(format string, v ...any)  { l.logOutput(LogLevelFlagInfo, format, v...) }
func (l *zerologLogger) Warnf(format string, v ...any)  { l.logOutput(LogLevelFlagWarn, format, v...) }
func (l *zerologLogger) Errorf(format string, v ...any) { l.logOutput(LogLevelFlagError, format, v...) }
func (l *zerologLogger) Panicf(format string, v ...any) { l.logOutput(LogLevelFlagPanic, format, v...) }
func (l *zerologLogger) Fatalf(format string, v ...any) { l.logOutput(LogLevelFlagFatal, format, v...) }
