package httpx

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/middleware"
	"github.com/hewen/mastiff-go/middleware/recovery"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGinHandlerBuilder(t *testing.T) {
	initRoute := func(r *gin.Engine) {
		r.GET("/test", func(c *gin.Context) {
			ctx := contextkeys.ContextFrom(c)
			l := logger.NewLoggerWithContext(ctx)
			l.Infof("test")

			c.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		})
	}
	assert.NotNil(t, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(t, err)

	enableMetrics := true
	conf := &serverconf.HTTPConfig{
		Addr:         fmt.Sprintf("localhost:%d", port),
		PprofEnabled: true,
		Mode:         "debug",
		Middlewares: middleware.Config{
			EnableMetrics: &enableMetrics,
		},
	}

	builder := &GinHandlerBuilder{
		Conf:             conf,
		InitRoute:        initRoute,
		ExtraMiddlewares: []gin.HandlerFunc{recovery.GinMiddleware()},
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()

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
