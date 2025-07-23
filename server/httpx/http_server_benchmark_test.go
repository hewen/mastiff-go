package httpx

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/loggerconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

func BenchmarkParallelFiber(b *testing.B) {
	port := initFrameworkType(b, serverconf.FrameworkFiber)

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runClient(port)
		}
	})
}

func BenchmarkParallelGin(b *testing.B) {
	port := initFrameworkType(b, serverconf.FrameworkGin)
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runClient(port)
		}
	})
}

func BenchmarkFiber(b *testing.B) {
	port := initFrameworkType(b, serverconf.FrameworkFiber)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runClient(port)
	}
}

func BenchmarkGin(b *testing.B) {
	port := initFrameworkType(b, serverconf.FrameworkGin)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runClient(port)
	}
}

func initFrameworkType(b *testing.B, ft serverconf.FrameworkType) int {
	port, err := util.GetFreePort()
	assert.Nil(b, err)

	conf := &serverconf.HTTPConfig{
		Addr:          fmt.Sprintf("localhost:%d", port),
		Mode:          "debug",
		FrameworkType: ft,
	}

	s, err := NewHTTPServer(conf)
	assert.Nil(b, err)

	s.Get("/test", func(c unicontext.UniversalContext) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ok",
		})
	})

	go func() {
		defer s.Stop()
		s.Start()
	}()
	err = setLogger()
	assert.Nil(b, err)

	time.Sleep(100 * time.Millisecond)
	return port
}

func runClient(port int) {
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
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
