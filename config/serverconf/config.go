// Package serverconf provides configuration for HTTP, gRPC, and queue servers.
package serverconf

import (
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf"
)

type (
	// FrameworkType represents the type of framework used for the HTTP server.
	FrameworkType string

	// HTTPConfig holds the configuration for an HTTP server.
	HTTPConfig struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middlewareconf.Config
		// Addr represents the HTTP server address.
		Addr string
		// Mode represents the server mode, either "debug", "release", or "test".
		Mode string
		// FrameworkType either "gin", "fiber".
		FrameworkType FrameworkType
		// TimeoutRead represents the timeout for reading requests in milliseconds.
		ReadTimeout int64
		// TimeoutWrite represents the timeout for writing responses in milliseconds.
		WriteTimeout int64
		// IdleTimeout represents the timeout for idle responses in milliseconds.
		IdleTimeout int64
		// PprofEnabled represents whether to enable pprof for profiling.
		PprofEnabled bool
		// EnableMetrics represents whether to enable metrics.
		EnableMetrics bool
	}

	// RPCConfig holds the configuration for a gRPC server.
	RPCConfig struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middlewareconf.Config
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

const (
	// FrameworkGin represents the type of framework used for the HTTP server, which is Gin.
	FrameworkGin FrameworkType = "gin"
	// FrameworkFiber represents the type of framework used for the HTTP server, which is Fiber.
	FrameworkFiber FrameworkType = "fiber"
)
