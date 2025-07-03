// Package server config
package server

import "time"

type (
	// HTTPConfig holds the configuration for an HTTP server.
	HTTPConfig struct {
		Addr         string // HTTP server address
		TimeoutRead  int64  // Timeout for reading requests in milliseconds
		TimeoutWrite int64  // Timeout for writing responses in milliseconds
		PprofEnabled bool   // Enable pprof for profiling
		Mode         string // "debug", "release", or "test"
	}

	// GrpcConfig holds the configuration for a gRPC server.
	GrpcConfig struct {
		Addr       string // gRPC server address
		Timeout    int64  // Timeout for gRPC requests in milliseconds
		Reflection bool   // Enable gRPC reflection
	}

	// QueueConfig holds the configuration for a queue server.
	QueueConfig struct {
		PoolSize           int           // Size of the goroutine pool for processing queue messages
		EmptySleepInterval time.Duration // Duration to sleep when the queue is empty
	}
)
