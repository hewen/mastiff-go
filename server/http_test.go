package server

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := &HTTPConf{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
	}

	handler := func(_ *gin.Engine) {}

	s, err := NewHTTPServer(c, handler)
	assert.Nil(t, err)

	go func() {
		s.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := &HTTPConf{
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
	c := &HTTPConf{
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
