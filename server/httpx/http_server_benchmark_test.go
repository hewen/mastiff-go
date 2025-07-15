package httpx

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/hewen/mastiff-go/config/serverconf"
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
		Mode: "release",
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

	time.Sleep(100 * time.Millisecond)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
			assert.Nil(b, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			assert.Nil(b, err)

			assert.Equal(b, http.StatusOK, resp.StatusCode)
			assert.Contains(b, string(body), `"message":"ok"`)

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

	time.Sleep(100 * time.Millisecond)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
			assert.Nil(b, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			assert.Nil(b, err)

			assert.Equal(b, http.StatusOK, resp.StatusCode)
			assert.Contains(b, string(body), `"message":"ok"`)
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
		Mode: "release",
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

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
		assert.Nil(b, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		assert.Nil(b, err)

		assert.Equal(b, http.StatusOK, resp.StatusCode)
		assert.Contains(b, string(body), `"message":"ok"`)

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

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
		assert.Nil(b, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		assert.Nil(b, err)

		assert.Equal(b, http.StatusOK, resp.StatusCode)
		assert.Contains(b, string(body), `"message":"ok"`)
	}
}
