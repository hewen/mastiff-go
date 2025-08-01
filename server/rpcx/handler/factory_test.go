package handler

import (
	"net/http"
	"testing"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewHandler_Connect(t *testing.T) {
	s, err := NewHandler(&serverconf.RPCConfig{
		FrameworkType: serverconf.FrameworkConnect,
	}, RPCBuildParams{})
	assert.Nil(t, s)
	assert.NotNil(t, err)

	s, err = NewHandler(&serverconf.RPCConfig{
		FrameworkType: serverconf.FrameworkConnect,
	}, RPCBuildParams{
		ConnectRegisterMux: func(*http.ServeMux) {},
	})
	assert.NotNil(t, s)
	assert.Nil(t, err)
}

func TestNewHandler_Grpc(t *testing.T) {
	s, err := NewHandler(&serverconf.RPCConfig{
		FrameworkType: serverconf.FrameworkGrpc,
	}, RPCBuildParams{})
	assert.Nil(t, s)
	assert.NotNil(t, err)

	s, err = NewHandler(&serverconf.RPCConfig{
		FrameworkType: serverconf.FrameworkGrpc,
	}, RPCBuildParams{
		GrpcRegisterFunc: func(*grpc.Server) {},
	})
	assert.NotNil(t, s)
	assert.Nil(t, err)
}

func TestNewHandler_ErrorType(t *testing.T) {
	s, err := NewHandler(&serverconf.RPCConfig{}, RPCBuildParams{})
	assert.Nil(t, s)
	assert.NotNil(t, err)
}

func TestNewHandler_ErrorConfig(t *testing.T) {
	s, err := NewHandler(nil, RPCBuildParams{})
	assert.Nil(t, s)
	assert.NotNil(t, err)
}
