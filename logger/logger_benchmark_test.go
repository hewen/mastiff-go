package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLoggerStd(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "std",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	l := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("Benchmark std logger test message #%d", i)
	}
}

func BenchmarkLoggerZap(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "zap",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	l := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("Benchmark zap logger test message #%d", i)
	}
}

func BenchmarkLoggerZerolog(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "zerolog",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	l := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("Benchmark zer logger test message #%d", i)
	}
}

func BenchmarkLoggerParallelStd(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "std",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	logger := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof("Benchmark std logger test message")
		}
	})
}

func BenchmarkLoggerParallelZap(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "zap",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	logger := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof("Benchmark zap logger test message")
		}
	})
}

func BenchmarkLoggerParallelZerolog(b *testing.B) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "tmp.log")
	assert.Nil(b, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	err = InitLogger(Config{
		Backend: "zerolog",
		Outputs: []string{"file"},
		FileOutput: &FileOutputConfig{
			Path: tmpFile.Name(),
		},
	})
	assert.Nil(b, err)

	logger := NewLoggerWithTraceID("BENCHMARK_TRACE_ID")

	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof("Benchmark zer logger test message")
		}
	})
}
