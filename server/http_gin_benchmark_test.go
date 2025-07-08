package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGinHttpStdLogger(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "std",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	b.ResetTimer()
	fmt.Println("")
	for i := 0; i < b.N; i++ {
		testGinLogger()
	}
}

func BenchmarkGinHttpZapLogger(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "zap",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	b.ResetTimer()
	fmt.Println("")
	for i := 0; i < b.N; i++ {
		testGinLogger()
	}
}

func BenchmarkGinHttpZerologLogger(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "zerolog",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	b.ResetTimer()
	fmt.Println("")
	for i := 0; i < b.N; i++ {
		testGinLogger()
	}
}

func BenchmarkGinHttpStdLoggerParallel(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "std",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	fmt.Println("")
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testGinLogger()
		}
	})
}

func BenchmarkGinHttpZapLoggerParallel(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "zap",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	fmt.Println("")
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testGinLogger()
		}
	})
}

func BenchmarkGinHttpZerologLoggerParallel(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = logger.InitLogger(logger.Config{
		Backend: "zerolog",
		Output:  tmpFile.Name(),
	})
	assert.Nil(b, err)

	fmt.Println("")
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testGinLogger()
		}
	})
}

func testGinLogger() {
	handler := GinLoggerHandler()
	gin.SetMode(gin.ReleaseMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/test", bytes.NewReader([]byte("{'test':1}")))
	ctx.Request.Header.Add("Content-Type", "application/json")
	ctx.Request.Header.Set("User-Agent", "GIN-GO-SERVER")

	handler(ctx)
}
