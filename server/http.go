package server

import (
	"context"
	"net/http"
	"time"

	"github.com/hewen/mastiff-go/logger"
)

const (
	HTTPTimeoutReadDefault  = 10
	HTTPTimeoutWriteDefault = 10
)

type HTTPService struct {
	addr string
	s    *http.Server
	l    *logger.Logger
}

func NewHTTPServer(conf HTTPConfig, handler http.Handler) (*HTTPService, error) {
	if conf.TimeoutRead == 0 {
		conf.TimeoutRead = HTTPTimeoutReadDefault
	}
	if conf.TimeoutWrite == 0 {
		conf.TimeoutWrite = HTTPTimeoutWriteDefault
	}

	srv := &http.Server{
		Addr:         conf.Addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.TimeoutRead) * time.Second,
		WriteTimeout: time.Duration(conf.TimeoutWrite) * time.Second,
	}

	return &HTTPService{
		addr: conf.Addr,
		s:    srv,
		l:    logger.NewLogger(),
	}, nil
}

func (s *HTTPService) Start() {
	AddGracefulStop(s.Stop)
	gracefulStop()

	s.l.Infof("Start http service %s", s.addr)
	if err := s.s.ListenAndServe(); err != http.ErrServerClosed {
		s.l.Errorf("http service failed: %v", err)
	}
}

func (s *HTTPService) Stop() {
	s.l.Infof("Shutdown service %s", s.addr)
	if s.s == nil {
		return
	}

	err := s.s.Shutdown(context.Background())
	if err != nil {
		s.l.Errorf("%v", err)
	}
	s.s = nil
}
