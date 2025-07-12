package server

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := &serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
	}

	handler := func(_ *gin.Engine) {}

	s, err := NewHTTPServer(c, handler)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())

	var server Servers
	server.Add(s)
	go func() {
		defer server.Stop()
		server.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	handler := func(_ *gin.Engine) {}
	s, err := NewHTTPServer(c, handler)
	assert.Nil(t, err)

	s.Stop()

	s.s = nil
	s.Stop()
}

func TestHTTPServerStartError(t *testing.T) {
	c := &serverconf.HTTPConfig{
		Addr: "error addr",
	}

	handler := func(_ *gin.Engine) {}
	s, err := NewHTTPServer(c, handler)
	assert.Nil(t, err)

	s.Start()
}

func TestHTTPServerEmptyConfig(t *testing.T) {
	handler := func(_ *gin.Engine) {}
	_, err := NewHTTPServer(nil, handler)
	assert.EqualValues(t, err, ErrEmptyHTTPConf)
}
