// Package server provides configuration for HTTP, gRPC, and queue servers.
package server

import (
	"time"

	"github.com/hewen/mastiff-go/middleware"
)

type (
	// HTTPConf holds the configuration for an HTTP server.
	HTTPConf struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middleware.Config
		// Addr represents the HTTP server address.
		Addr string
		// Mode represents the server mode, either "debug", "release", or "test".
		Mode string
		// TimeoutRead represents the timeout for reading requests in milliseconds.
		TimeoutRead int64
		// TimeoutWrite represents the timeout for writing responses in milliseconds.
		TimeoutWrite int64
		// PprofEnabled represents whether to enable pprof for profiling.
		PprofEnabled bool
	}

	// GrpcConf holds the configuration for a gRPC server.
	GrpcConf struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middleware.Config
		// Addr represents the gRPC server address.
		Addr string
		// Timeout represents the timeout for gRPC requests in milliseconds.
		Timeout int64
		// Reflection represents whether to enable gRPC reflection.
		Reflection bool
	}

	// QueueConf holds the configuration for a queue server.
	QueueConf struct {
		// QueueName represents the queue name.
		QueueName string
		// PoolSize represents the size of the goroutine pool for processing queue messages.
		PoolSize int
		// EmptySleepInterval represents the duration to sleep when the queue is empty.
		EmptySleepInterval time.Duration
	}
)
