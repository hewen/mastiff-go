// Package server config
package server

type (
	// HTTPConfig holds the configuration for an HTTP server.
	HTTPConfig struct {
		Addr         string
		TimeoutRead  int64
		TimeoutWrite int64
		PprofEnabled bool
		Mode         string // "debug", "release", or "test"
	}

	// GrpcConfig holds the configuration for a gRPC server.
	GrpcConfig struct {
		Addr       string
		Timeout    int64
		Reflection bool
	}
)
