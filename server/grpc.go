// Package server grpc server implementation
package server

import (
	"errors"
	"net"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware/logging"
	"github.com/hewen/mastiff-go/middleware/recovery"
	"github.com/hewen/mastiff-go/middleware/timeout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	// ErrEmptyGrpcConf is returned when the gRPC configuration is empty.
	ErrEmptyGrpcConf = errors.New("empty grpc config")

	// ErrGrpcExecPanic is an error that indicates a panic occurred during gRPC execution.
	ErrGrpcExecPanic = errors.New("grpc exec panic")
)

const (
	defaultTimeout = 30
)

// GrpcServer represents a gRPC server.
type GrpcServer struct {
	addr string
	s    *grpc.Server
	l    logger.Logger
	ln   net.Listener
	mu   sync.Mutex
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(conf *GrpcConf, registerServerFunc func(*grpc.Server), interceptors ...grpc.UnaryServerInterceptor) (*GrpcServer, error) {
	if conf == nil {
		return nil, ErrEmptyGrpcConf
	}
	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		return nil, err
	}

	srv := &GrpcServer{
		addr: conf.Addr,
		l:    logger.NewLogger(),
		ln:   ln,
	}

	if conf.Timeout == 0 {
		conf.Timeout = defaultTimeout
	}

	var serverInterceptors []grpc.UnaryServerInterceptor

	serverInterceptors = append(serverInterceptors,
		timeout.UnaryServerInterceptor(time.Duration(conf.Timeout)*time.Second),
		recovery.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(),
	)

	if len(interceptors) > 0 {
		serverInterceptors = append(serverInterceptors, interceptors...)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			serverInterceptors...,
		)),
	}

	srv.s = grpc.NewServer(opts...)

	registerServerFunc(srv.s)

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
