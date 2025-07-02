// Package logger config
package logger

// Config holds the configuration for the logger.
type Config struct {
	Level   LogLevel
	Output  string
	MaxSize int
}
