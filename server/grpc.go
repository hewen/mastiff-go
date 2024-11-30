package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	defaultTimeout = 30
)

var ErrGrpcExecPanic = errors.New("grpc exec panic")

type GrpcServer struct {
	addr string
	s    *grpc.Server
	l    *logger.Logger
	ln   net.Listener
}

func NewGrpcServer(conf GrpcConfig, registerServerFunc func(*grpc.Server), interceptors ...grpc.UnaryServerInterceptor) (*GrpcServer, error) {
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

	return srv, nil
}

func (s *GrpcServer) Start() {
	AddGracefulStop(s.Stop)
	gracefulStop()

	s.l.Infof("Start grpc service %s", s.addr)
	if err := s.s.Serve(s.ln); err != nil {
		s.l.Errorf("grpc service failed: %v", err)
	}
}

func (s *GrpcServer) Stop() {
	s.l.Infof("Shutdown grpc service %s", s.addr)
	if s.s != nil {
		s.s.GracefulStop()
		s.s = nil
	}

	if s.ln != nil {
		s.ln.Close()
		s.ln = nil
	}
}

func (s *GrpcServer) middleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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

func (s *GrpcServer) execHandler(ctx context.Context, req interface{}, handler grpc.UnaryHandler, l *logger.Logger) (data interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrGrpcExecPanic
			l.Errorf("%v $%s", req, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
		}
	}()
	return handler(ctx, req)
}

func (s *GrpcServer) timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()

		return handler(ctx, req)
	}
}
