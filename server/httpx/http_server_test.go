package httpx

import (
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}
	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)

	s.WithLogger(logger.NewLogger())
	assert.Equal(t, fmt.Sprintf(`http std server(%s)`, conf.Addr), s.Name())

	go func() {
		defer s.Stop()
		s.Start()
	}()
}

func TestHTTPServerStop(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)

	s.Stop()
}

func TestHTTPServerStartError(t *testing.T) {
	conf := &serverconf.HTTPConfig{
		Addr: "error addr",
	}

	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)

	s.Start()
}

func TestHTTPServerEmptyConfig(t *testing.T) {
	initRoute := func(_ *gin.Engine) {}
	builder := &GinHandlerBuilder{
		Conf:      nil,
		InitRoute: initRoute,
	}
	_, err := NewHTTPServer(builder)
	assert.EqualValues(t, err, ErrEmptyHTTPConf)
}

func TestToDuration(t *testing.T) {
	dur := toDuration(0)
	assert.Equal(t, HTTPTimeoutDefault*time.Second, dur)

	dur = toDuration(1)
	assert.Equal(t, 1*time.Second, dur)
}
