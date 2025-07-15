// Package serverconf provides configuration for HTTP, gRPC, and queue servers.
package serverconf

import (
	"time"

	"github.com/hewen/mastiff-go/middleware"
)

type (
	// HTTPConfig holds the configuration for an HTTP server.
	HTTPConfig struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middleware.Config
		// Addr represents the HTTP server address.
		Addr string
		// Mode represents the server mode, either "debug", "release", or "test".
		Mode string
		// TimeoutRead represents the timeout for reading requests in milliseconds.
		ReadTimeout int64
		// TimeoutWrite represents the timeout for writing responses in milliseconds.
		WriteTimeout int64
		// PprofEnabled represents whether to enable pprof for profiling.
		PprofEnabled bool
	}

	// RPCConfig holds the configuration for a gRPC server.
	RPCConfig struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middleware.Config
		// Addr represents the gRPC server address.
		Addr string
		// Timeout represents the timeout for gRPC requests in milliseconds.
		Timeout int64
		// Reflection represents whether to enable gRPC reflection.
		Reflection bool
	}

	// QueueConfig holds the configuration for a queue server.
	QueueConfig struct {
		// QueueName represents the queue name.
		QueueName string
		// PoolSize represents the size of the goroutine pool for processing queue messages.
		PoolSize int
		// EmptySleepInterval represents the duration to sleep when the queue is empty.
		EmptySleepInterval time.Duration
	}
)
