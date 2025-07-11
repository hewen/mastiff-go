// Package server grpc server implementation
package server

import (
	"errors"
	"fmt"
	"net"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	// ErrEmptyGrpcConf is returned when the gRPC configuration is empty.
	ErrEmptyGrpcConf = errors.New("empty grpc config")
)

// GrpcServer represents a gRPC server.
type GrpcServer struct {
	s      *grpc.Server
	logger logger.Logger
	ln     net.Listener
	addr   string
	mu     sync.Mutex
}

// NewGrpcServer creates and initializes a new gRPC server with configured middlewares.
func NewGrpcServer(
	conf *GrpcConf,
	registerServerFunc func(*grpc.Server),
	extraInterceptors ...grpc.UnaryServerInterceptor,
) (*GrpcServer, error) {
	if conf == nil {
		return nil, ErrEmptyGrpcConf
	}

	// Initialize GrpcServer struct
	srv := &GrpcServer{
		addr:   conf.Addr,
		logger: logger.NewLogger(),
	}

	// Start listening on configured address
	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		srv.logger.Errorf("failed to listen on %s: %v", conf.Addr, err)
		return nil, err
	}
	srv.ln = ln

	// Load built-in middleware based on configuration
	interceptors := middleware.LoadGRPCMiddlewares(conf.Middlewares)

	// Append extra user-defined interceptors
	interceptors = append(interceptors, extraInterceptors...)

	// Apply chained interceptors
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
	}

	srv.s = grpc.NewServer(opts...)

	// Register application-specific services
	registerServerFunc(srv.s)

	// Enable reflection if configured
	if conf.Reflection {
		reflection.Register(srv.s)
	}

	return srv, nil
}

// Start starts the gRPC server and listens for incoming connections.
func (s *GrpcServer) Start() {
	if err := s.s.Serve(s.ln); err != nil {
		s.logger.Errorf("grpc service failed: %v", err)
	}
}

// Stop gracefully stops the gRPC server and closes the listener.
func (s *GrpcServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.s != nil {
		s.s.GracefulStop()
		s.s = nil
	}

	if s.ln != nil {
		err := s.ln.Close()
		if err != nil {
			s.logger.Errorf("%v", err)
		}
		s.ln = nil
	}
}

// Name returns the name of the gRPC server.
func (s *GrpcServer) Name() string {
	return fmt.Sprintf("gRPC server(%s)", s.addr)
}

// WithLogger sets the logger for the gRPC server.
func (s *GrpcServer) WithLogger(l logger.Logger) {
	if l != nil {
		s.logger = l
	}
}
