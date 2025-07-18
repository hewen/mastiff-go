package server

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
	"github.com/stretchr/testify/assert"
)

func TestNewGinAPIHandler(t *testing.T) {
	handler := func(r *gin.Engine) {
		r.GET("/test", func(c *gin.Context) {
			l := logger.NewLoggerWithContext(c.Request.Context())
			l.Infof("test")

			c.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		})
	}
	assert.NotNil(t, handler)

	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &HTTPConf{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
		Mode:         "debug",
	}

	server, err := NewHTTPServer(conf, handler)
	assert.Nil(t, err)

	go func() {
		server.Start()
	}()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
	assert.Nil(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `"message":"ok"`)
}
