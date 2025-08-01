// Package server provides a logging server implementation.
package server

import (
	"github.com/hewen/mastiff-go/logger"
)

// LoggingServer wraps a Server and logs start and stop events. It is used to provide logging for server implementations.
type LoggingServer struct {
	// Inner represents the inner server.
	Inner Server
	// Logger represents the logger for logging start and stop events.
	Logger logger.Logger
}

// Name returns the name of the inner server.
func (s *LoggingServer) Name() string {
	return s.Inner.Name()
}

// Start starts the inner server and logs the event.
func (s *LoggingServer) Start() {
	s.Logger.Infof("[server] starting %s", s.Inner.Name())
	s.Inner.Start()
}

// Stop stops the inner server and logs the event.
func (s *LoggingServer) Stop() {
	s.Logger.Infof("[server] stopping %s", s.Inner.Name())
	s.Inner.Stop()
	s.Logger.Infof("[server] stopped %s", s.Inner.Name())
}

// WithLogger sets the logger for the logging server.
func (s *LoggingServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.Logger = l
	}
}
