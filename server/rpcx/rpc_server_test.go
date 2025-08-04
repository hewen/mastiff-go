package rpcx

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/rpcx/handler"
	"github.com/stretchr/testify/assert"
)

type MockRPCHandler struct{}

func (*MockRPCHandler) Start() error {
	return fmt.Errorf("start error")
}

func (*MockRPCHandler) Stop() error {
	return nil
}

func (*MockRPCHandler) Name() string {
	return "mock rpc"
}

func TestRPCServer(t *testing.T) {
	s, err := NewRPCServer(&serverconf.RPCConfig{
		FrameworkType: serverconf.FrameworkConnect,
	}, handler.RPCBuildParams{
		ConnectRegisterMux: func(*http.ServeMux) {},
	})
	assert.NotNil(t, s)
	assert.Nil(t, err)
}

func TestRPCServerError(t *testing.T) {
	s, err := NewRPCServer(nil, handler.RPCBuildParams{
		ConnectRegisterMux: func(*http.ServeMux) {},
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
}

func TestRPCServerStart(t *testing.T) {
	s := &RPCServer{
		handler: &MockRPCHandler{},
		logger:  logger.NewLogger(),
	}

	assert.Panics(t, func() {
		s.Start()
	})

	s.Stop()
	s.WithLogger(logger.NewLogger())
	s.Name()
}
