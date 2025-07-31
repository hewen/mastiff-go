// Package serverconf provides configuration for HTTP, gRPC, and queue servers.
package serverconf

import (
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/panjf2000/gnet/v2"
)

const (
	defaultTickInterval = 1 * time.Minute
)

type (
	// HTTPFrameworkType represents the type of framework used for the HTTP server.
	HTTPFrameworkType string

	// RPCFrameworkType represents the type of framework used for the RPC server.
	RPCFrameworkType string

	// SocketFrameworkType represents the type of framework used for the RPC server.
	SocketFrameworkType string

	// HTTPConfig holds the configuration for an HTTP server.
	HTTPConfig struct {
		// Middlewares represents the configuration for middlewares.
		Middlewares middlewareconf.Config
		// Addr represents the HTTP server address.
		Addr string
		// Mode represents the server mode, either "debug", "release", or "test".
		Mode string
		// FrameworkType either "gin", "fiber".
		FrameworkType HTTPFrameworkType
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
		// FrameworkType either "grpc", "connect".
		FrameworkType RPCFrameworkType
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

	// SocketConfig holds the configuration for a socket server.
	SocketConfig struct {
		// Addr represents the socket server address.
		Addr string
		// FrameworkType either "gnet".
		FrameworkType SocketFrameworkType
		// GnetOptions represents the options for the gnet framework.
		GnetOptions gnet.Options
		// TickInterval represents the interval for the tick function.
		TickInterval time.Duration
	}
)

const (
	// FrameworkGin represents the type of framework used for the HTTP server, which is Gin.
	FrameworkGin HTTPFrameworkType = "gin"
	// FrameworkFiber represents the type of framework used for the HTTP server, which is Fiber.
	FrameworkFiber HTTPFrameworkType = "fiber"

	// FrameworkGrpc represents the type of framework used for the rpc server, which is gRPC.
	FrameworkGrpc RPCFrameworkType = "grpc"
	// FrameworkConnect represents the type of framework used for the rpc server, which is connect.
	FrameworkConnect RPCFrameworkType = "connect"

	// FrameworkGnet represents the type of framework used for the socket server, which is gnet.
	FrameworkGnet SocketFrameworkType = "gnet"
)

// SetDefault sets default values for the configuration.
func (c *SocketConfig) SetDefault() {
	if c.TickInterval == 0 {
		c.TickInterval = defaultTickInterval
	}
}
