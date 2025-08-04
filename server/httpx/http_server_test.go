package httpx

import (
	"fmt"
	"testing"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/httpx/handler"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr:          fmt.Sprintf("localhost:%d", port),
		PprofEnabled:  true,
		FrameworkType: serverconf.FrameworkGin,
	}

	s, err := NewHTTPServer(conf)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())
	assert.Equal(t, fmt.Sprintf(`http gin server(%s)`, conf.Addr), s.Name())

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr:          fmt.Sprintf("localhost:%d", port),
		FrameworkType: serverconf.FrameworkGin,
	}

	s, err := NewHTTPServer(conf)
	assert.Nil(t, err)

	s.Stop()
}

func TestHTTPServerStartError(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr:          "error addr",
		FrameworkType: serverconf.FrameworkGin,
	}

	s, err := NewHTTPServer(conf)
	assert.Nil(t, err)

	assert.Panics(t, func() {
		s.Start()
	})
}

func TestHTTPServerEmptyConfig(t *testing.T) {
	_, err := NewHTTPServer(nil)
	assert.EqualValues(t, err, handler.ErrEmptyHTTPConf)
}
