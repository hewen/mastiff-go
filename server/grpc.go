// Package server grpc server implementation
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

var (
	// ErrEmptyGrpcConfig is returned when the gRPC configuration is empty.
	ErrEmptyGrpcConfig = errors.New("empty grpc config")

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
	l    *logger.Logger
	ln   net.Listener
	mu   sync.Mutex
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(conf *GrpcConfig, registerServerFunc func(*grpc.Server), interceptors ...grpc.UnaryServerInterceptor) (*GrpcServer, error) {
	if conf == nil {
		return nil, ErrEmptyGrpcConfig
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

	serverInterceptors = append(serverInterceptors, srv.timeoutInterceptor(time.Duration(conf.Timeout)*time.Second), srv.middleware)
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

// middleware is a gRPC interceptor that logs the request and response details, including execution time and any errors.
func (s *GrpcServer) middleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	begin := time.Now()
	pr, ok := peer.FromContext(ctx)
	var addr string
	if ok && pr != nil {
		addr = (pr.Addr).(*net.TCPAddr).IP.String()
	}

	ctx = logger.NewOutgoingContextWithIncomingContext(ctx)
	l := logger.NewLoggerWithContext(ctx)

	resp, err = s.execHandler(ctx, req, handler, l)

	var errStr string
	if err != nil {
		errStr = fmt.Sprintf(" | err: %s", err.Error())
	}

	switch {
	case errStr != "":
		l.Errorf("%10s | %15s | %-10s | %v | %v%s", util.FormatDuration(time.Since(begin)), addr, info.FullMethod, req, resp, errStr)
	case time.Since(begin) > time.Second:
		l.Infof("SLOW %10s | %15s | %-10s | %v | %v%s", util.FormatDuration(time.Since(begin)), addr, info.FullMethod, req, resp, errStr)
	default:
		l.Infof("%10s | %15s | %-10s | %v | %v%s", util.FormatDuration(time.Since(begin)), addr, info.FullMethod, req, resp, errStr)
	}

	return resp, err
}

// execHandler executes the handler and recovers from any panic, logging the error if it occurs.
func (s *GrpcServer) execHandler(ctx context.Context, req any, handler grpc.UnaryHandler, l *logger.Logger) (data any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrGrpcExecPanic
			l.Errorf("%v $%s", req, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
		}
	}()
	return handler(ctx, req)
}

// timeoutInterceptor creates a gRPC interceptor that sets a timeout for each request.
func (s *GrpcServer) timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()

		return handler(ctx, req)
	}
}
