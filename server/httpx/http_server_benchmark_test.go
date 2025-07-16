package httpx

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/loggerconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func BenchmarkParallelFiber(b *testing.B) {
	initRoute := func(app *fiber.App) {
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(map[string]string{
				"message": "ok",
			})
		})

	}
	assert.NotNil(b, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(b, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Mode: "debug",
	}

	builder := &FiberHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(b, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()
	err = setLogger()
	assert.Nil(b, err)

	time.Sleep(100 * time.Millisecond)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}
		}
	})
}

func BenchmarkParallelGin(b *testing.B) {
	initRoute := func(r *gin.Engine) {
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		})
	}
	assert.NotNil(b, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(b, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Mode: "release",
	}

	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(b, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()

	err = setLogger()
	assert.Nil(b, err)

	time.Sleep(100 * time.Millisecond)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}
		}
	})
}

func BenchmarkFiber(b *testing.B) {
	initRoute := func(app *fiber.App) {
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(map[string]string{
				"message": "ok",
			})
		})

	}
	assert.NotNil(b, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(b, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Mode: "debug",
	}

	builder := &FiberHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(b, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()

	err = setLogger()
	assert.Nil(b, err)

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
		if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}
	}
}

func BenchmarkGin(b *testing.B) {
	initRoute := func(r *gin.Engine) {
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		})
	}
	assert.NotNil(b, initRoute)

	port, err := util.GetFreePort()
	assert.Nil(b, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
		Mode: "release",
	}

	builder := &GinHandlerBuilder{
		Conf:      conf,
		InitRoute: initRoute,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(b, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()

	err = setLogger()
	assert.Nil(b, err)

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
		if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}
	}
}

func setLogger() error {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	return logger.InitLogger(loggerconf.Config{
		Backend: "zerolog",
		Outputs: []string{"file"},
		FileOutput: &loggerconf.FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
}
