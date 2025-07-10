// Package server config
package server

import (
	"time"

	"github.com/hewen/mastiff-go/middleware"
)

type (
	// HTTPConf holds the configuration for an HTTP server.
	HTTPConf struct {
		Addr         string // HTTP server address
		TimeoutRead  int64  // Timeout for reading requests in milliseconds
		TimeoutWrite int64  // Timeout for writing responses in milliseconds
		PprofEnabled bool   // Enable pprof for profiling
		Mode         string // "debug", "release", or "test"
		Middlewares  middleware.Config
	}

	// GrpcConf holds the configuration for a gRPC server.
	GrpcConf struct {
		Addr        string // gRPC server address
		Timeout     int64  // Timeout for gRPC requests in milliseconds
		Reflection  bool   // Enable gRPC reflection
		Middlewares middleware.Config
	}

	// QueueConf holds the configuration for a queue server.
	QueueConf struct {
		PoolSize           int           // Size of the goroutine pool for processing queue messages
		EmptySleepInterval time.Duration // Duration to sleep when the queue is empty
		QueueName          string        // queue name
	}
)
