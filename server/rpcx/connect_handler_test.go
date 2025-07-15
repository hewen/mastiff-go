package rpcx

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestConnectServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)
	enableMetrics := true
	c := &serverconf.RPCConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Middlewares: middleware.Config{
			EnableMetrics: &enableMetrics,
		},
		Reflection: true,
	}

	builder := &ConnectHandlerBuilder{
		Conf:        c,
		RegisterMux: func(_ *http.ServeMux) {},
	}

	s, err := NewRPCServer(builder)
	assert.NotNil(t, s)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())
	_ = s.Name()

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

func TestConnectHandler_StartError(t *testing.T) {
	connectServer := &ConnectHandler{
		server: &http.Server{
			ReadTimeout: time.Second,
		},
		ln:   &brokenListener{},
		addr: "mock",
	}

	err := connectServer.Start()
	assert.NotNil(t, err)
}

func TestConnectServerStop(t *testing.T) {
	c := &serverconf.RPCConfig{}

	builder := &ConnectHandlerBuilder{
		Conf:        c,
		RegisterMux: func(_ *http.ServeMux) {},
	}

	s, err := NewRPCServer(builder)
	assert.Nil(t, err)
	assert.NotNil(t, s)

	s.Stop()
}

func TestConnectServerEmptyConfig(t *testing.T) {
	builder := &ConnectHandlerBuilder{
		Conf:        nil,
		RegisterMux: func(_ *http.ServeMux) {},
	}

	_, err := NewRPCServer(builder)
	assert.EqualValues(t, err, ErrEmptyRPCConf)
}

func TestNewConnectServerError(t *testing.T) {
	c := &serverconf.RPCConfig{
		Addr: "error",
	}

	builder := &ConnectHandlerBuilder{
		Conf:        c,
		RegisterMux: func(_ *http.ServeMux) {},
	}

	_, err := NewRPCServer(builder)
	assert.EqualValues(t, "listen tcp: address error: missing port in address", err.Error())
}
