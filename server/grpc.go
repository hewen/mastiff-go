// Package server grpc server implementation
package server

import (
	"errors"
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
	addr string
	s    *grpc.Server
	l    logger.Logger
	ln   net.Listener
	mu   sync.Mutex
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

	// Start listening on configured address
	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		logger.NewLogger().Errorf("failed to listen on %s: %v", conf.Addr, err)
		return nil, err
	}

	// Initialize GrpcServer struct
	srv := &GrpcServer{
		addr: conf.Addr,
		ln:   ln,
		l:    logger.NewLogger(),
	}

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
	AddGracefulStop(s.Stop)
	gracefulStop()

	s.l.Infof("Start grpc service %s", s.addr)
	if err := s.s.Serve(s.ln); err != nil {
		s.l.Errorf("grpc service failed: %v", err)
	}
}

// Stop gracefully stops the gRPC server and closes the listener.
func (s *GrpcServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.l.Infof("Shutdown grpc service %s", s.addr)
	if s.s != nil {
		s.s.GracefulStop()
		s.s = nil
	}

	if s.ln != nil {
		err := s.ln.Close()
		if err != nil {
			s.l.Errorf("%v", err)
		}
		s.ln = nil
	}
}
