package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
)

func TestNewGinAPIHandler(t *testing.T) {
	handler := NewGinAPIHandler(func(r *gin.Engine) {
		r.GET("/test")
	})
	assert.NotNil(t, handler)

	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	_, err = NewHTTPServer(conf, handler)
	assert.Nil(t, err)
}

func TestGinLoggerHandler(_ *testing.T) {
	handler := GinLoggerHandler()
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/test", bytes.NewReader([]byte("{'test':1}")))
	ctx.Request.Header.Add("Content-Type", "application/json")

	handler(ctx)
}

func TestGinRecoverHandler(t *testing.T) {
	handler := NewGinAPIHandler(func(r *gin.Engine) {
		r.GET("/test", func(_ *gin.Context) {
			panic("test")
		})
	})
	assert.NotNil(t, handler)

	port, err := util.GetFreePort()
	assert.Nil(t, err)

	addr := fmt.Sprintf("localhost:%d", port)
	conf := HTTPConfig{
		Addr: addr,
	}

	s, err := NewHTTPServer(conf, handler)
	assert.Nil(t, err)

	go func() {
		s.Start()
	}()

	for {
		res, err := http.Get("http://" + addr + "/test")
		if res == nil || res.StatusCode != http.StatusOK {
			continue
		}
		defer func() {
			_ = res.Body.Close()
		}()
		assert.Nil(t, err)
		return
	}
}
