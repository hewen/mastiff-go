// Package server config
package server

import (
	"time"

	"github.com/hewen/mastiff-go/middleware"
)

type (
	// HTTPConf holds the configuration for an HTTP server.
	HTTPConf struct {
		Middlewares  middleware.Config
		Addr         string // HTTP server address
		Mode         string // "debug", "release", or "test"
		TimeoutRead  int64  // Timeout for reading requests in milliseconds
		TimeoutWrite int64  // Timeout for writing responses in milliseconds
		PprofEnabled bool   // Enable pprof for profiling
	}

	// GrpcConf holds the configuration for a gRPC server.
	GrpcConf struct {
		Middlewares middleware.Config
		Addr        string // gRPC server address
		Timeout     int64  // Timeout for gRPC requests in milliseconds
		Reflection  bool   // Enable gRPC reflection
	}

	// QueueConf holds the configuration for a queue server.
	QueueConf struct {
		QueueName          string        // queue name
		PoolSize           int           // Size of the goroutine pool for processing queue messages
		EmptySleepInterval time.Duration // Duration to sleep when the queue is empty
	}
)
