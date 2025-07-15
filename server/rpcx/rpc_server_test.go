package rpcx

import (
	"fmt"
	"testing"

	"github.com/hewen/mastiff-go/logger"
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

type MockRPCHandlerBuilder struct {
}

func (*MockRPCHandlerBuilder) BuildRPC() (RPCHandler, error) {
	return &MockRPCHandler{}, nil
}

type MockRPCHandlerBuilderError struct {
}

func (*MockRPCHandlerBuilderError) BuildRPC() (RPCHandler, error) {
	return &MockRPCHandler{}, fmt.Errorf("error")
}

func TestRPCServer(t *testing.T) {
	builder := &MockRPCHandlerBuilder{}

	s, err := NewRPCServer(builder)
	assert.NotNil(t, s)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())
	_ = s.Name()
	s.Start()
	s.Stop()
}

func TestRPCServerError(t *testing.T) {
	builder := &MockRPCHandlerBuilderError{}

	s, err := NewRPCServer(builder)
	assert.Nil(t, s)
	assert.NotNil(t, err)

}
