package handler

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestConnectServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)
	enableMetrics := true
	c := &serverconf.RPCConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Middlewares: middlewareconf.Config{
			EnableMetrics: &enableMetrics,
		},
		Reflection: true,
	}

	s, err := NewConnectHandler(c, func(_ *http.ServeMux) {})
	assert.NotNil(t, s)
	assert.Nil(t, err)

	_ = s.Name()

	go func() {
		defer func() { _ = s.Stop() }()
		_ = s.Start()
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

	s, err := NewConnectHandler(c, func(_ *http.ServeMux) {})
	assert.Nil(t, err)
	assert.NotNil(t, s)

	err = s.Stop()
	assert.Nil(t, err)
}

func TestConnectServerEmptyConfig(t *testing.T) {
	_, err := NewConnectHandler(
		nil,
		func(_ *http.ServeMux) {},
	)

	assert.EqualValues(t, err, ErrEmptyRPCConf)
}

func TestNewConnectServerError(t *testing.T) {
	c := &serverconf.RPCConfig{
		Addr: "error",
	}

	_, err := NewConnectHandler(
		c,
		func(_ *http.ServeMux) {},
	)

	assert.EqualValues(t, "listen tcp: address error: missing port in address", err.Error())
}
