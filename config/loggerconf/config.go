// Package loggerconf provides a configurable logger.
package loggerconf

import (
	"errors"
	"fmt"
	"slices"
)

// LogLevel represents the log level as a string.
type LogLevel string

// Config defines logger configuration.
type Config struct {
	// Level defines the logging level, e.g. "INFO", "DEBUG".
	Level LogLevel
	// FileOutput defines file output configuration (used when "file" is in Outputs).
	FileOutput *FileOutputConfig
	// Backend specifies the logging backend: "std", "zap", or "zerolog".
	Backend string
	// Outputs defines the output targets, e.g. []string{"stdout", "file"}
	Outputs []string
	// EnableMasking enables masking of sensitive fields in log entries.
	EnableMasking bool
}

// FileOutputConfig defines configuration for file output.
type FileOutputConfig struct {
	// Path defines the file path for the log file.
	Path string
	// RotatePolicy defines the file rotation policy: "daily", "size", or "none".
	RotatePolicy string
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	MaxSize int
	// Compress defines whether to compress rotated log files.
	Compress bool
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	if slices.Contains(cfg.Outputs, "file") && (cfg.FileOutput == nil || cfg.FileOutput.Path == "") {
		return errors.New("file output selected but FileOutput.Path is empty")
	}
	if cfg.Backend != "zap" && cfg.Backend != "zerolog" && cfg.Backend != "std" && cfg.Backend != "" {
		return fmt.Errorf("unsupported backend: %s", cfg.Backend)
	}
	return nil
}
