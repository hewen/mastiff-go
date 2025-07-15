package httpx

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}
	s, err := NewHTTPServer(&conf, builder)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())
	assert.Equal(t, fmt.Sprintf(`http server(%s)`, conf.Addr), s.Name())

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(&conf, builder)
	assert.Nil(t, err)

	s.Stop()

	s.s = nil
	s.Stop()
}

func TestHTTPServerStartError(t *testing.T) {
	conf := serverconf.HTTPConfig{
		Addr: "error addr",
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(&conf, builder)
	assert.Nil(t, err)

	s.Start()
}

func TestHTTPServerEmptyConfig(t *testing.T) {
	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      serverconf.HTTPConfig{},
		InitRoute: initRoute,
	}
	_, err := NewHTTPServer(nil, builder)
	assert.EqualValues(t, err, ErrEmptyHTTPConf)
}
