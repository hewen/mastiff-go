// Package logger config
package logger

// Config defines logger configuration.
type Config struct {
	// Level defines the logging level, e.g. "INFO", "DEBUG".
	Level LogLevel
	// Output defines the file path for log output; if empty, logs will not be written to file.
	Output string
	// MaxSize defines max log file size in megabytes before rotating.
	MaxSize int
	// Backend specifies the logging backend: "std", "zap", or "zerolog".
	Backend string
}
