package server

import (
	"fmt"
	"testing"

	"github.com/hewen/mastiff-go/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	s, err := NewHTTPServer(c, gin.New())
	assert.Nil(t, err)

	go func() {
		s.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	c := HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	s, err := NewHTTPServer(c, gin.New())
	assert.Nil(t, err)

	s.Stop()

	s.s = nil
	s.Stop()
}

func TestHTTPServerStartError(t *testing.T) {
	c := HTTPConfig{
		Addr: "error addr",
	}

	s, err := NewHTTPServer(c, gin.New())
	assert.Nil(t, err)

	s.Start()
}
