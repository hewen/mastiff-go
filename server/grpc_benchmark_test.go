package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func BenchmarkGrpcStdLogger(b *testing.B) {
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
		testLoggerInterceptor()
	}
}

func BenchmarkGrpcZapLogger(b *testing.B) {
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
		testLoggerInterceptor()
	}
}

func BenchmarkGrpcZerologLogger(b *testing.B) {
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
		testLoggerInterceptor()
	}
}

func BenchmarkGrpcStdLoggerParallel(b *testing.B) {
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
			testLoggerInterceptor()
		}
	})
}

func BenchmarkGrpcZapLoggerParallel(b *testing.B) {
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
			testLoggerInterceptor()
		}
	})
}

func BenchmarkGrpcZerologLoggerParallel(b *testing.B) {
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
			testLoggerInterceptor()
		}
	})
}

func testLoggerInterceptor() {
	gs := GrpcServer{}
	_, _ = gs.loggerInterceptor(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})
}
